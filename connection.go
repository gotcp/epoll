package epoll

import (
	"net"

	"golang.org/x/sys/unix"
)

func (ep *EP) Add(fd int) error {
	var err = unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLET, Fd: int32(fd)})
	if err != nil {
		return err
	}
	unix.SetNonblock(fd, true)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, ep.KeepAlive)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, ep.ReadBuffer)
	return nil
}

func (ep *EP) Del(fd int) error {
	var err error
	if err = unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}
	if err = unix.Close(fd); err != nil {
		return err
	}
	return nil
}

func (ep *EP) AddConn(conn net.Conn) (int, error) {
	var fd = GetConnFd(conn)
	return fd, ep.Add(fd)
}

func (ep *EP) CloseConn(conn net.Conn) {
	var fd = GetConnFd(conn)
	ep.CloseAction(fd)
}
