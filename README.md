# Golang high-performance asynchronous TCP using Epoll (only supports Linux)

## Example
 

```go
package main

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/gotcp/epoll"
)

// Asynchronous event, when using Start
func OnAccept(fd int) {
	fmt.Printf("OnAccept -> %d\n", fd)
}

// Asynchronous event, when using Start1
func OnAccept1(fd int, conn net.Conn) {
	fmt.Printf("OnAccept1 -> %d\n", fd)
}

// Asynchronous event
func OnReceive(fd int, msg []byte, n int) {
	// var err = epoll.Write(fd, msg)
	var err = epoll.WriteWithTimeout(fd, msg, 3*time.Second)
	if err != nil {
		fmt.Printf("OnReceive -> %d, %v\n", fd, err)
	}
}

// Asynchronous event
func OnClose(fd int) {
	fmt.Printf("OnClose -> %d\n", fd)
}

// Asynchronous event
func OnError(fd int, code epoll.ErrorCode, err error) {
	if fd > 0 && code == epoll.ERROR_CLOSE_CONNECTION {
		fmt.Printf("OnError -> %d, %d, %v\n", fd, code, err)
	} else {
		fmt.Printf("OnError -> %d, %v\n", code, err)
	}
}

var ep *epoll.EP

func main() {
	var err error

	var rLimit syscall.Rlimit
	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// parameters: host, port, readBuffer, numberOfThreads, maxQueueLength
	ep, err = epoll.New(1024, 3000, 100000)
	if err != nil {
		panic(err)
	}
	defer ep.Stop()

	// one of OnAccept or OnAccept1, it depends on the use of ep.Start or ep.Start1
	ep.OnAccept = OnAccept   // optional
	ep.OnReceive = OnReceive // must have, when using Start
	ep.OnClose = OnClose     // optional
	ep.OnError = OnError     // optional
	// ep.OnAccept1 = OnAccept1 // when using Start1

	// use pure EPOLL
	ep.Start("127.0.0.1", 8001)
}
```