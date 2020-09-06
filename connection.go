package epoll

import (
	"errors"
	"fmt"
	"time"

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
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, ep.WriteBuffer)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, ep.ReadTimeout)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, ep.WriteTimeout)
	return err
}

func (ep *EP) Delete(fd int) error {
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func (ep *EP) CloseFd(fd int) error {
	return unix.Close(fd)
}

// called externally
func (ep *EP) EstablishConnection(fd int) error {
	var sequenceId = ep.GetSequenceId()
	ep.AddConnection(fd, sequenceId)
	var err error
	if err = ep.Add(fd); err != nil {
		ep.DeleteConnection(fd)
	}
	return err
}

// called externally
func (ep *EP) DestroyConnection(fd int) error {
	var err error
	var sequenceId = ep.GetConnectionSequenceId(fd)
	if sequenceId < 0 {
		err = errors.New(fmt.Sprintf("%d not found in the list", fd))
		return err
	}
	if err = ep.Delete(fd); err == nil {
		ep.InvokeClose(sequenceId, fd)
	}
	return err
}

func (ep *EP) Close(fd int) error {
	var err = ep.Delete(fd)
	if err == nil {
		ep.DeleteConnection(fd)
		ep.CloseFd(fd)
	}
	return err
}

func (ep *EP) CloseAll() {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var fd int
	for fd, _ = range ep.Connections.List {
		if ep.Delete(fd) == nil {
			delete(ep.Connections.List, fd)
			ep.CloseFd(fd)
		}
	}
}

func (ep *EP) DeleteConnection(fd int) {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	delete(ep.Connections.List, fd)
}

func (ep *EP) AddConnection(fd int, sequenceId int) {
	var conn = &Conn{
		Fd:         fd,
		SequenceId: sequenceId,
		Timestamp:  time.Now().Unix(),
	}
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	ep.Connections.List[fd] = conn
}

func (ep *EP) GetConnectionSequenceId(fd int) int {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Timestamp = time.Now().Unix()
		return conn.SequenceId
	}
	return -1
}

func (ep *EP) UpdateConnection(fd int) {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Timestamp = time.Now().Unix()
	}
}
