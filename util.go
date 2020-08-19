package epoll

import (
	"context"
	"net"
	"reflect"
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

func GetConnFd(conn net.Conn) int {
	var tcpConn = reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	var fdVal = tcpConn.FieldByName("fd")
	var pfdVal = reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
