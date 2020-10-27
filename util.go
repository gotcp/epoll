package epoll

import (
	"context"
	"time"

	"golang.org/x/sys/unix"
)

func Write(fd int, msg []byte) (int, error) {
	return unix.Write(fd, msg)
}

func WriteWithTimeout(fd int, msg []byte, timeout time.Duration) (int, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var done = make(chan int8, 1)

	var err error
	var writed int

	go func() {
		writed, err = Write(fd, msg)
		done <- 1
	}()

	select {
	case <-done:
		return writed, err
	case <-ctx.Done():
		return -1, ctx.Err()
	}
}

func (ep *EP) WriteSSL(fd int, msg []byte, n int) (int, int) {
	var ssl = ep.GetConnectionSSL(fd)
	if ssl != nil {
		return sslWrite(ssl.SSL, msg, n)
	}
	return -1, -1
}

func (ep *EP) WriteSSLWithTimeout(fd int, msg []byte, n int, timeout time.Duration) (int, int) {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var done = make(chan int8)

	var writed, errno int

	go func() {
		writed, errno = ep.WriteSSL(fd, msg, n)
		done <- 1
	}()

	select {
	case <-done:
		return writed, errno
	case <-ctx.Done():
		return -1, SSL_ERROR_TIMEOUT
	}
}
