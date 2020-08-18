package epoll

import (
	"golang.org/x/sys/unix"
)

func Write(fd int, msg []byte) error {
	var _, err = unix.Write(fd, msg)
	return err
}
