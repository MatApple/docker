package main
 
import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"docker"
	"rcli"
	"term"
	"os"
	"os/signal"
	"syscall"
)
 
var localAddr *string = flag.String("l", "0.0.0.0:7000", "local address")
var remoteAddr *string = flag.String("r", "127.0.0.1:4242", "remote address")
 

func runRemoteCommand(args []string) error {
	// basically a modified version of dockers own runCommand....
	
	if conn, err := rcli.Call("tcp", "127.0.0.1:4242", args...); err == nil {
		options := conn.GetOptions()
		if options.RawTerminal &&
			term.IsTerminal(int(os.Stdin.Fd())) &&
			os.Getenv("NORAW") == "" {
			if oldState, err := rcli.SetRawTerminal(); err != nil {
				return err
			} else {
				defer rcli.RestoreTerminal(oldState)
			}
		}
		receiveStdout := docker.Go(func() error {
			_, err := io.Copy(os.Stdout, conn)
			return err
		})
		sendStdin := docker.Go(func() error {
			_, err := io.Copy(conn, os.Stdin)
			if err := conn.CloseWrite(); err != nil {
				log.Printf("Couldn't send EOF: " + err.Error())
			}
			return err
		})
		if err := <-receiveStdout; err != nil {
			return err
		}
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			if err := <-sendStdin; err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Can't connect to docker daemon. Is 'docker -d' running on this host?")
	}
	return nil
}




func proxyConn(conn *net.TCPConn) {
	n, err := conn.Read(data)
	if err != nil {
		panic(err)
		
	}
	if err := runRemoteCommand(n); err != nil {
			log.Fatal(err)
		}
}
 
func handleConn(in <-chan *net.TCPConn, out chan<- *net.TCPConn) {
	for conn := range in {
		proxyConn(conn)
		out <- conn
	}
}
 
func closeConn(in <-chan *net.TCPConn) {
	for conn := range in {
		conn.Close()
	}
}
 
func main() {
	flag.Parse()
 
	fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)
 
	addr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		panic(err)
	}
 
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
 
	pending, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)
 
	for i := 0; i < 5; i++ {
		go handleConn(pending, complete)
	}
	go closeConn(complete)
 
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		pending <- conn
	}
}