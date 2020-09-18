package main

import (
	"C"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
	"unsafe"

	"libguestfs.org/libnbd"
	"libguestfs.org/nbdkit"
)

const (
	pidPath = "/tmp/nbd-redirect.pid"
	socketPath = "/tmp/nbd-redirect.sock"
	startupTimeoutSeconds = 5
)

type FakeVddk struct {
	nbdkit.Plugin
}

type FakeVddkConnection struct {
	nbdkit.Connection
	Server *exec.Cmd
	Client *libnbd.Libnbd
}

func (p *FakeVddk) CanMultiConn() (bool, error) {
	return false, nil
}

func (p *FakeVddk) CanWrite() (bool, error) {
	return false, nil
}

func (p *FakeVddk) Config(key string, value string) error {
	return nil
}

func (p *FakeVddk) ConfigComplete() error {
	return nil
}

func (p *FakeVddk) Open(readonly bool) (nbdkit.ConnectionInterface, error) {
	fmt.Println("Opening")
	args := []string{
		"--foreground",
		"--readonly",
		"--exit-with-parent",
		"--unix", socketPath,
		"--pidfile", pidPath,
		"--filter", "/opt/testing/nbdkit-xz-filter.so",
		"file", "file=/opt/testing/nbdtest.xz",
	}

	server := exec.Command("nbdkit", args...)
	stdout, _ := server.StdoutPipe()
	stderr, _ := server.StderrPipe()
	err := server.Start()
	if err != nil {
		return nil, err
	}

	err = WaitForNbd(pidPath)
	if err != nil {
		stdoutb, _ := ioutil.ReadAll(stdout)
		stderrb, _ := ioutil.ReadAll(stderr)
		message := fmt.Sprintf("Error starting nbdkit: %s\nOutput: %s\nError: %s", err.Error(), string(stdoutb), string(stderrb))
		return nil, nbdkit.PluginError{Errmsg: message}
	}

	client, err := libnbd.Create()
	if err != nil {
		return nil, err
	}

	err = client.ConnectUri(fmt.Sprintf("nbd+unix://?socket=%s", socketPath))
	if err != nil {
		return nil, err
	}

	return &FakeVddkConnection{
		Server: server,
		Client: client,
	}, nil
}

func (c *FakeVddkConnection) Close() {
	c.Server.Process.Kill()
	c.Client.Close()
	os.Remove(pidPath)
	os.Remove(socketPath)
}

func (c *FakeVddkConnection) GetSize() (uint64, error) {
	return c.Client.GetSize()
}

// WaitForNbd waits for nbdkit to start by watching for the existence of the given PID file.
func WaitForNbd(pidfile string) error {
	nbdCheck := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-nbdCheck:
				return
			case <-time.After(500 * time.Millisecond):
				_, err := os.Stat(pidfile)
				if err == nil {
					nbdCheck <- true
					return
				}
			}
		}
	}()

	select {
	case <-nbdCheck:
		return nil
	case <-time.After(startupTimeoutSeconds * time.Second):
		nbdCheck <- true
		return nbdkit.PluginError{Errmsg: "timed out waiting for nbdkit to be ready"}
	}
}

func (c *FakeVddkConnection) PRead(buf []byte, offset uint64, flags uint32) error {
	err := c.Client.Pread(buf, offset, nil)
	if err != nil {
		return err
	}

	return nil
}

//export plugin_init
func plugin_init() unsafe.Pointer {
	return nbdkit.PluginInitialize("vddk", &FakeVddk{})
}

func main() {}
