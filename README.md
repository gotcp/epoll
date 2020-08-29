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
func OnReceive(sequenceId int, fd int, msg []byte, n int) {
	// var err = epoll.Write(fd, msg)
	var err = epoll.WriteWithTimeout(fd, msg, 3*time.Second)
	if err != nil {
		fmt.Printf("OnReceive -> %d, %v\n", fd, err)
	}
}

// Synchronous event. This event will be triggered before closing fd
func OnClose(sequenceId int, fd int) {
	fmt.Printf("OnClose -> %d\n", fd)
}

// Asynchronous event
func OnError(sequenceId int, fd int, code epoll.ErrorCode, err error) {
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

	// parameters: readBuffer, threads, queueLength
	ep, err = epoll.New(2048, 3000, 4096)
	if err != nil {
		panic(err)
	}
	defer ep.Stop()

	ep.OnReceive = OnReceive // must have
	ep.OnError = OnError     // optional
	ep.OnAccept = OnAccept   // optional
	ep.OnClose = OnClose     // optional

	ep.Start("127.0.0.1", 8001)
}
```