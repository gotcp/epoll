package epoll

import (
	"context"
	"time"

	"golang.org/x/sys/unix"
)

func Write(fd int, msg []byte) error {
	var _, err = unix.Write(fd, msg)
	return err
}

func WriteWithTimeout(fd int, msg []byte, timeout time.Duration) error {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var done = make(chan int8, 1)
	var err error
	go func() {
		err = Write(fd, msg)
		done <- 1
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
