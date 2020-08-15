# Golang high-performance asynchronous TCP using Epoll (only supports Linux)

## Example
 

```go
package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/gotcp/epoll"
)

// Asynchronous event
func OnAccept(fd int) {
	fmt.Printf("OnAccept -> %d\n", fd)
}

// Asynchronous event
func OnReceive(msg []byte, fd int) {
	var err = epoll.Write(fd, msg)
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

var server *epoll.Server

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
	server, err = epoll.New("127.0.0.1", 8001, 1024, 3000, 100000)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	server.SetTimeout(int(5 * time.Second))

	server.OnReceive = OnReceive // must have
	server.OnAccept = OnAccept   // optional
	server.OnClose = OnClose     // optional
	server.OnError = OnError     // optional

	server.Start()
}
```