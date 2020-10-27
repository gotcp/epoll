package epoll

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

const (
	EPOLL_EVENTS          = unix.EPOLLET
	EPOLL_EVENTS_EPOLLIN  = unix.EPOLLIN | unix.EPOLLET
	EPOLL_EVENTS_EPOLLOUT = unix.EPOLLIN | unix.EPOLLOUT | unix.EPOLLET
)

type UpdateDataFunc func(data interface{})

func connRecycleUpdate(ptr interface{}) {
	var conn, ok = ptr.(*Conn)
	if ok && conn != nil {
		resetConn(conn)
	}
}

func resetConn(conn *Conn) {
	conn.Fd = -1
	conn.SSL = nil
	conn.Data = nil
	conn.SequenceId = -1
	conn.Timestamp = 0
	conn.Status = 0
}

func (ep *EP) getConn() *Conn {
	var conn, err = ep.connPool.Get()
	if err == nil {
		return conn.(*Conn)
	}
	return nil
}

func (ep *EP) putConn(conn *Conn) {
	resetConn(conn)
	ep.connPool.PutWithId(conn, conn.Id)
}

func (ep *EP) Add(fd int) error {
	var event = &unix.EpollEvent{
		Events: EPOLL_EVENTS_EPOLLIN,
		Fd:     int32(fd),
	}
	var err = unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_ADD, fd, event)
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

func (ep *EP) EnableEpollIn(fd int) error {
	var event = &unix.EpollEvent{
		Events: EPOLL_EVENTS_EPOLLIN,
		Fd:     int32(fd),
	}
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_MOD, fd, event)
}

func (ep *EP) DisableEpollIn(fd int) error {
	var event = &unix.EpollEvent{
		Events: EPOLL_EVENTS,
		Fd:     int32(fd),
	}
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_MOD, fd, event)
}

func (ep *EP) EnableEpollOut(fd int) error {
	var event = &unix.EpollEvent{
		Events: EPOLL_EVENTS_EPOLLOUT,
		Fd:     int32(fd),
	}
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_MOD, fd, event)
}

func (ep *EP) DisableEpollOut(fd int) error {
	var event = &unix.EpollEvent{
		Events: EPOLL_EVENTS_EPOLLIN,
		Fd:     int32(fd),
	}
	return unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_MOD, fd, event)
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
		err = errors.New(fmt.Sprintf(ErrorTemplateNotFound, fd))
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
	var fd int
	var conn *Conn
	var ok1, ok2 bool
	ep.Connections.IterateAndUpdate(func(key interface{}, value interface{}) bool {
		fd, ok1 = key.(int)
		conn, ok2 = value.(*Conn)
		if ok1 && ok2 {
			ep.Delete(fd)
			ep.putConnSSL(conn)
			ep.putConn(conn)
			ep.CloseFd(fd)
		}
		return false
	})
	if ep.IsSSL {
		cMallocTrim()
	}
}

func (ep *EP) DeleteConnection(fd int) bool {
	var c *Conn
	var ok bool
	ep.Connections.RemoveAndUpdate(fd, func(value interface{}) {
		c, ok = value.(*Conn)
		if ok {
			ep.putConnSSL(c)
			ep.putConn(c)
		}
	})
	return ok
}

func (ep *EP) AddConnection(fd int, sequenceId int) {
	var conn = ep.getConn()
	conn.Fd = fd
	conn.SequenceId = sequenceId
	conn.Data = nil
	conn.Timestamp = time.Now().Unix()
	conn.Status = 0
	ep.Connections.Put(fd, conn)
}

func (ep *EP) GetConnectionSequenceId(fd int) int {
	var sequenceId = -1
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
			sequenceId = c.SequenceId
		}
	})
	return sequenceId
}

func (ep *EP) GetConnection(fd int) *Conn {
	var c *Conn
	var ok bool
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
		}
	})
	return c
}

func (ep *EP) GetConnectionAndSequenceId(fd int) (*Conn, int) {
	var c *Conn
	var sequenceId = -1
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
			sequenceId = c.SequenceId
		}
	})
	return c, sequenceId
}

func (ep *EP) UpdateConnection(fd int) {
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
		}
	})
}

func (ep *EP) SetConnectionStatus(fd int, status int) bool {
	var c *Conn
	var ok bool
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		c, ok = value.(*Conn)
		if ok {
			c.Status = status
		}
	})
	return ok
}

func (ep *EP) GetConnectionStatus(fd int) (int, bool) {
	var value = ep.Connections.Get(fd)
	var c, ok = value.(*Conn)
	if ok {
		return c.Status, ok
	}
	return -1, ok
}

func (ep *EP) GetConnectionCount() int {
	return ep.Connections.GetCount()
}

func (ep *EP) SetConnectionData(fd int, data interface{}) bool {
	var c *Conn
	var ok bool
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		c, ok = value.(*Conn)
		if ok {
			c.Data = data
		}
	})
	return ok
}

func (ep *EP) GetConnectionData(fd int) (interface{}, bool) {
	var value = ep.Connections.Get(fd)
	var c, ok = value.(*Conn)
	if ok {
		return c.Data, ok
	}
	return nil, ok
}

func (ep *EP) UpdateConnectionDataWithFunc(fd int, updateDataFunc UpdateDataFunc) bool {
	var c *Conn
	var ok bool
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		c, ok = value.(*Conn)
		if ok {
			updateDataFunc(c.Data)
		}
	})
	return ok
}
