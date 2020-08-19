package epoll

import (
	"net"
	"reflect"

	"golang.org/x/sys/unix"
)

func Write(fd int, msg []byte) error {
	var _, err = unix.Write(fd, msg)
	return err
}

func GetConnFd(conn net.Conn) int {
	var tcpConn = reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	var fdVal = tcpConn.FieldByName("fd")
	var pfdVal = reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
