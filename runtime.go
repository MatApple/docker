package docker

import (
	"container/list"
	"fmt"
	"github.com/dotcloud/docker/auth"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"
)

type Runtime struct {
	root           string
	repository     string
	containers     *list.List
	networkManager *NetworkManager
	graph          *Graph
	repositories   *TagStore
	authConfig     *auth.AuthConfig
	idIndex        *TruncIndex
}

var sysInitPath string

func init() {
	sysInitPath = SelfPath()
}

func (runtime *Runtime) List() []*Container {
	containers := new(History)
	for e := runtime.containers.Front(); e != nil; e = e.Next() {
		containers.Add(e.Value.(*Container))
	}
	return *containers
}

func (runtime *Runtime) getContainerElement(id string) *list.Element {
	for e := runtime.containers.Front(); e != nil; e = e.Next() {
		container := e.Value.(*Container)
		if container.Id == id {
			return e
		}
	}
	return nil
}

func (runtime *Runtime) Get(name string) *Container {
	id, err := runtime.idIndex.Get(name)
	if err != nil {
		return nil
	}
	e := runtime.getContainerElement(id)
	if e == nil {
		return nil
	}
	return e.Value.(*Container)
}

func (runtime *Runtime) Exists(id string) bool {
	return runtime.Get(id) != nil
}

func (runtime *Runtime) containerRoot(id string) string {
	return path.Join(runtime.repository, id)
}

func (runtime *Runtime) Create(config *Config) (*Container, error) {
	// Lookup image
	img, err := runtime.repositories.LookupImage(config.Image)
	if err != nil {
		return nil, err
	}
	// Generate id
	id := GenerateId()
	// Generate default hostname
	// FIXME: the lxc template no longer needs to set a default hostname
	if config.Hostname == "" {
		config.Hostname = id[:12]
	}

	container := &Container{
		// FIXME: we should generate the ID here instead of receiving it as an argument
		Id:              id,
		Created:         time.Now(),
		Path:            config.Cmd[0],
		Args:            config.Cmd[1:], //FIXME: de-duplicate from config
		Config:          config,
		Image:           img.Id, // Always use the resolved image id
		NetworkSettings: &NetworkSettings{},
		// FIXME: do we need to store this in the container?
		SysInitPath: sysInitPath,
	}
	container.root = runtime.containerRoot(container.Id)
	// Step 1: create the container directory.
	// This doubles as a barrier to avoid race conditions.
	if err := os.Mkdir(container.root, 0700); err != nil {
		return nil, err
	}

	// If custom dns exists, then create a resolv.conf for the container
	if len(config.Dns) > 0 {
		container.ResolvConfPath = path.Join(container.root, "resolv.conf")
		f, err := os.Create(container.ResolvConfPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		for _, dns := range config.Dns {
			if _, err := f.Write([]byte("nameserver " + dns + "\n")); err != nil {
				return nil, err
			}
		}
	} else {
		container.ResolvConfPath = "/etc/resolv.conf"
	}

	// Step 2: save the container json
	if err := container.ToDisk(); err != nil {
		return nil, err
	}
	// Step 3: register the container
	if err := runtime.Register(container); err != nil {
		return nil, err
	}
	return container, nil
}

func (runtime *Runtime) Load(id string) (*Container, error) {
	container := &Container{root: runtime.containerRoot(id)}
	if err := container.FromDisk(); err != nil {
		return nil, err
	}
	if container.Id != id {
		return container, fmt.Errorf("Container %s is stored at %s", container.Id, id)
	}
	if container.State.Running {
		container.State.Ghost = true
	}
	if err := runtime.Register(container); err != nil {
		return nil, err
	}
	return container, nil
}

// Register makes a container object usable by the runtime as <container.Id>
func (runtime *Runtime) Register(container *Container) error {
	if container.runtime != nil || runtime.Exists(container.Id) {
		return fmt.Errorf("Container is already loaded")
	}
	if err := validateId(container.Id); err != nil {
		return err
	}

	// init the wait lock
	container.waitLock = make(chan struct{})

	// FIXME: if the container is supposed to be running but is not, auto restart it?
	//        if so, then we need to restart monitor and init a new lock
	// If the container is supposed to be running, make sure of it
	if container.State.Running {
		if output, err := exec.Command("lxc-info", "-n", container.Id).CombinedOutput(); err != nil {
			return err
		} else {
			if !strings.Contains(string(output), "RUNNING") {
				Debugf("Container %s was supposed to be running be is not.", container.Id)
				container.State.setStopped(-127)
				if err := container.ToDisk(); err != nil {
					return err
				}
			}
		}
	}

	// If the container is not running or just has been flagged not running
	// then close the wait lock chan (will be reset upon start)
	if !container.State.Running {
		close(container.waitLock)
	}

	// Even if not running, we init the lock (prevents races in start/stop/kill)
	container.State.initLock()

	container.runtime = runtime

	// Attach to stdout and stderr
	container.stderr = newWriteBroadcaster()
	container.stdout = newWriteBroadcaster()
	// Attach to stdin
	if container.Config.OpenStdin {
		container.stdin, container.stdinPipe = io.Pipe()
	} else {
		container.stdinPipe = NopWriteCloser(ioutil.Discard) // Silently drop stdin
	}
	// done
	runtime.containers.PushBack(container)
	runtime.idIndex.Add(container.Id)
	return nil
}

func (runtime *Runtime) LogToDisk(src *writeBroadcaster, dst string) error {
	log, err := os.OpenFile(dst, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	src.AddWriter(log)
	return nil
}

func (runtime *Runtime) Destroy(container *Container) error {
	element := runtime.getContainerElement(container.Id)
	if element == nil {
		return fmt.Errorf("Container %v not found - maybe it was already destroyed?", container.Id)
	}

	if err := container.Stop(); err != nil {
		return err
	}
	if mounted, err := container.Mounted(); err != nil {
		return err
	} else if mounted {
		if err := container.Unmount(); err != nil {
			return fmt.Errorf("Unable to unmount container %v: %v", container.Id, err)
		}
	}
	// Deregister the container before removing its directory, to avoid race conditions
	runtime.idIndex.Delete(container.Id)
	runtime.containers.Remove(element)
	if err := os.RemoveAll(container.root); err != nil {
		return fmt.Errorf("Unable to remove filesystem for %v: %v", container.Id, err)
	}
	return nil
}

// Commit creates a new filesystem image from the current state of a container.
// The image can optionally be tagged into a repository
func (runtime *Runtime) Commit(id, repository, tag, comment string) (*Image, error) {
	container := runtime.Get(id)
	if container == nil {
		return nil, fmt.Errorf("No such container: %s", id)
	}
	// FIXME: freeze the container before copying it to avoid data corruption?
	// FIXME: this shouldn't be in commands.
	rwTar, err := container.ExportRw()
	if err != nil {
		return nil, err
	}
	// Create a new image from the container's base layers + a new layer from container changes
	img, err := runtime.graph.Create(rwTar, container, comment)
	if err != nil {
		return nil, err
	}
	// Register the image if needed
	if repository != "" {
		if err := runtime.repositories.Set(repository, tag, img.Id, true); err != nil {
			return img, err
		}
	}
	return img, nil
}

func (runtime *Runtime) restore() error {
	dir, err := ioutil.ReadDir(runtime.repository)
	if err != nil {
		return err
	}
	for _, v := range dir {
		id := v.Name()
		container, err := runtime.Load(id)
		if err != nil {
			Debugf("Failed to load container %v: %v", id, err)
			continue
		}
		Debugf("Loaded container %v", container.Id)
	}
	return nil
}

// FIXME: harmonize with NewGraph()
func NewRuntime() (*Runtime, error) {
	return NewRuntimeFromDirectory("/var/lib/docker")
}

func NewRuntimeFromDirectory(root string) (*Runtime, error) {
	runtimeRepo := path.Join(root, "containers")

	if err := os.MkdirAll(runtimeRepo, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	g, err := NewGraph(path.Join(root, "graph"))
	if err != nil {
		return nil, err
	}
	repositories, err := NewTagStore(path.Join(root, "repositories"), g)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create Tag store: %s", err)
	}
	if NetworkBridgeIface == "" {
		NetworkBridgeIface = DefaultNetworkBridge
	}
	netManager, err := newNetworkManager(NetworkBridgeIface)
	if err != nil {
		return nil, err
	}
	authConfig, err := auth.LoadConfig(root)
	if err != nil && authConfig == nil {
		// If the auth file does not exist, keep going
		return nil, err
	}
	runtime := &Runtime{
		root:           root,
		repository:     runtimeRepo,
		containers:     list.New(),
		networkManager: netManager,
		graph:          g,
		repositories:   repositories,
		authConfig:     authConfig,
		idIndex:        NewTruncIndex(),
	}

	if err := runtime.restore(); err != nil {
		return nil, err
	}
	return runtime, nil
}

type History []*Container

func (history *History) Len() int {
	return len(*history)
}

func (history *History) Less(i, j int) bool {
	containers := *history
	return containers[j].When().Before(containers[i].When())
}

func (history *History) Swap(i, j int) {
	containers := *history
	tmp := containers[i]
	containers[i] = containers[j]
	containers[j] = tmp
}

func (history *History) Add(container *Container) {
	*history = append(*history, container)
	sort.Sort(history)
}
