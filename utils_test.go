package docker

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
)

func TestBufReader(t *testing.T) {
	reader, writer := io.Pipe()
	bufreader := newBufReader(reader)

	// Write everything down to a Pipe
	// Usually, a pipe should block but because of the buffered reader,
	// the writes will go through
	done := make(chan bool)
	go func() {
		writer.Write([]byte("hello world"))
		writer.Close()
		done <- true
	}()

	// Drain the reader *after* everything has been written, just to verify
	// it is indeed buffering
	<-done
	output, err := ioutil.ReadAll(bufreader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, []byte("hello world")) {
		t.Error(string(output))
	}
}

type dummyWriter struct {
	buffer      bytes.Buffer
	failOnWrite bool
}

func (dw *dummyWriter) Write(p []byte) (n int, err error) {
	if dw.failOnWrite {
		return 0, errors.New("Fake fail")
	}
	return dw.buffer.Write(p)
}

func (dw *dummyWriter) String() string {
	return dw.buffer.String()
}

func (dw *dummyWriter) Close() error {
	return nil
}

func TestWriteBroadcaster(t *testing.T) {
	writer := newWriteBroadcaster()

	// Test 1: Both bufferA and bufferB should contain "foo"
	bufferA := &dummyWriter{}
	writer.AddWriter(bufferA)
	bufferB := &dummyWriter{}
	writer.AddWriter(bufferB)
	writer.Write([]byte("foo"))

	if bufferA.String() != "foo" {
		t.Errorf("Buffer contains %v", bufferA.String())
	}

	if bufferB.String() != "foo" {
		t.Errorf("Buffer contains %v", bufferB.String())
	}

	// Test2: bufferA and bufferB should contain "foobar",
	// while bufferC should only contain "bar"
	bufferC := &dummyWriter{}
	writer.AddWriter(bufferC)
	writer.Write([]byte("bar"))

	if bufferA.String() != "foobar" {
		t.Errorf("Buffer contains %v", bufferA.String())
	}

	if bufferB.String() != "foobar" {
		t.Errorf("Buffer contains %v", bufferB.String())
	}

	if bufferC.String() != "bar" {
		t.Errorf("Buffer contains %v", bufferC.String())
	}

	// Test3: Test removal
	writer.RemoveWriter(bufferB)
	writer.Write([]byte("42"))
	if bufferA.String() != "foobar42" {
		t.Errorf("Buffer contains %v", bufferA.String())
	}
	if bufferB.String() != "foobar" {
		t.Errorf("Buffer contains %v", bufferB.String())
	}
	if bufferC.String() != "bar42" {
		t.Errorf("Buffer contains %v", bufferC.String())
	}

	// Test4: Test eviction on failure
	bufferA.failOnWrite = true
	writer.Write([]byte("fail"))
	if bufferA.String() != "foobar42" {
		t.Errorf("Buffer contains %v", bufferA.String())
	}
	if bufferC.String() != "bar42fail" {
		t.Errorf("Buffer contains %v", bufferC.String())
	}
	// Even though we reset the flag, no more writes should go in there
	bufferA.failOnWrite = false
	writer.Write([]byte("test"))
	if bufferA.String() != "foobar42" {
		t.Errorf("Buffer contains %v", bufferA.String())
	}
	if bufferC.String() != "bar42failtest" {
		t.Errorf("Buffer contains %v", bufferC.String())
	}

	writer.CloseWriters()
}

type devNullCloser int

func (d devNullCloser) Close() error {
	return nil
}

func (d devNullCloser) Write(buf []byte) (int, error) {
	return len(buf), nil
}

// This test checks for races. It is only useful when run with the race detector.
func TestRaceWriteBroadcaster(t *testing.T) {
	writer := newWriteBroadcaster()
	c := make(chan bool)
	go func() {
		writer.AddWriter(devNullCloser(0))
		c <- true
	}()
	writer.Write([]byte("hello"))
	<-c
}

// Test the behavior of TruncIndex, an index for querying IDs from a non-conflicting prefix.
func TestTruncIndex(t *testing.T) {
	index := NewTruncIndex()
	// Get on an empty index
	if _, err := index.Get("foobar"); err == nil {
		t.Fatal("Get on an empty index should return an error")
	}

	// Spaces should be illegal in an id
	if err := index.Add("I have a space"); err == nil {
		t.Fatalf("Adding an id with ' ' should return an error")
	}

	id := "99b36c2c326ccc11e726eee6ee78a0baf166ef96"
	// Add an id
	if err := index.Add(id); err != nil {
		t.Fatal(err)
	}
	// Get a non-existing id
	assertIndexGet(t, index, "abracadabra", "", true)
	// Get the exact id
	assertIndexGet(t, index, id, id, false)
	// The first letter should match
	assertIndexGet(t, index, id[:1], id, false)
	// The first half should match
	assertIndexGet(t, index, id[:len(id)/2], id, false)
	// The second half should NOT match
	assertIndexGet(t, index, id[len(id)/2:], "", true)

	id2 := id[:6] + "blabla"
	// Add an id
	if err := index.Add(id2); err != nil {
		t.Fatal(err)
	}
	// Both exact IDs should work
	assertIndexGet(t, index, id, id, false)
	assertIndexGet(t, index, id2, id2, false)

	// 6 characters or less should conflict
	assertIndexGet(t, index, id[:6], "", true)
	assertIndexGet(t, index, id[:4], "", true)
	assertIndexGet(t, index, id[:1], "", true)

	// 7 characters should NOT conflict
	assertIndexGet(t, index, id[:7], id, false)
	assertIndexGet(t, index, id2[:7], id2, false)

	// Deleting a non-existing id should return an error
	if err := index.Delete("non-existing"); err == nil {
		t.Fatalf("Deleting a non-existing id should return an error")
	}

	// Deleting id2 should remove conflicts
	if err := index.Delete(id2); err != nil {
		t.Fatal(err)
	}
	// id2 should no longer work
	assertIndexGet(t, index, id2, "", true)
	assertIndexGet(t, index, id2[:7], "", true)
	assertIndexGet(t, index, id2[:11], "", true)

	// conflicts between id and id2 should be gone
	assertIndexGet(t, index, id[:6], id, false)
	assertIndexGet(t, index, id[:4], id, false)
	assertIndexGet(t, index, id[:1], id, false)

	// non-conflicting substrings should still not conflict
	assertIndexGet(t, index, id[:7], id, false)
	assertIndexGet(t, index, id[:15], id, false)
	assertIndexGet(t, index, id, id, false)
}

func assertIndexGet(t *testing.T, index *TruncIndex, input, expectedResult string, expectError bool) {
	if result, err := index.Get(input); err != nil && !expectError {
		t.Fatalf("Unexpected error getting '%s': %s", input, err)
	} else if err == nil && expectError {
		t.Fatalf("Getting '%s' should return an error", input)
	} else if result != expectedResult {
		t.Fatalf("Getting '%s' returned '%s' instead of '%s'", input, result, expectedResult)
	}
}

func assertKernelVersion(t *testing.T, a, b *KernelVersionInfo, result int) {
	if r := CompareKernelVersion(a, b); r != result {
		t.Fatalf("Unepected kernel version comparaison result. Found %d, expected %d", r, result)
	}
}

func TestCompareKernelVersion(t *testing.T) {
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0},
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0},
		0)
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 2, Major: 6, Minor: 0},
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0},
		-1)
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0},
		&KernelVersionInfo{Kernel: 2, Major: 6, Minor: 0},
		1)
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0, Flavor: "0"},
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0, Flavor: "16"},
		0)
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 5},
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0},
		1)
	assertKernelVersion(t,
		&KernelVersionInfo{Kernel: 3, Major: 0, Minor: 20, Flavor: "25"},
		&KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0, Flavor: "0"},
		-1)
}
