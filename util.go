package epoll

import (
	"context"
	"errors"
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

func (ep *EP) WriteSSL(fd int, msg []byte, n int) error {
	var err error
	var ssl = ep.GetConnectionSSL(fd)
	if ssl != nil {
		if sslWrite(ssl.SSL, msg, n) {
			return nil
		}
	} else {
		err = errors.New("fd not found in the list")
		return err
	}
	err = errors.New("SSL write error")
	return err
}

func (ep *EP) WriteSSLWithTimeout(fd int, msg []byte, n int, timeout time.Duration) error {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var done = make(chan int8, 1)
	var err error
	go func() {
		err = ep.WriteSSL(fd, msg, n)
		done <- 1
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
