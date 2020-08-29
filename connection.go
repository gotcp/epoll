package epoll

import (
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

func (ep *EP) Delete(fd int) error {
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func (ep *EP) Close(fd int) error {
	return unix.Close(fd)
}
