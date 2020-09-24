package epoll

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
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
	var conn *Conn
	for fd, conn = range ep.Connections.List {
		ep.Delete(fd)
		ep.putConnSSL(conn)
		ep.putConn(conn)
		delete(ep.Connections.List, fd)
		ep.Connections.Count--
		ep.CloseFd(fd)
	}
	if ep.IsSSL {
		cMallocTrim()
	}
}

func (ep *EP) DeleteConnection(fd int) bool {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		ep.putConnSSL(conn)
		ep.putConn(conn)
		delete(ep.Connections.List, fd)
		ep.Connections.Count--
		return true
	}
	return false
}

func (ep *EP) AddConnection(fd int, sequenceId int) {
	var conn = ep.getConn()
	conn.Fd = fd
	conn.SequenceId = sequenceId
	conn.Timestamp = time.Now().Unix()
	conn.Status = 0
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	ep.Connections.List[fd] = conn
	ep.Connections.Count++
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

func (ep *EP) GetConnection(fd int) *Conn {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Timestamp = time.Now().Unix()
		return conn
	}
	return nil
}

func (ep *EP) GetConnectionAndSequenceId(fd int) (*Conn, int) {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Timestamp = time.Now().Unix()
		return conn, conn.SequenceId
	}
	return nil, -1
}

func (ep *EP) UpdateConnection(fd int) {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Timestamp = time.Now().Unix()
	}
}

func (ep *EP) SetConnectionStatus(fd int, status int) bool {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var c, ok = ep.Connections.List[fd]
	if ok {
		c.Status = status
	}
	return ok
}

func (ep *EP) GetConnectionStatus(fd int) (int, bool) {
	ep.Connections.Lock.RLock()
	defer ep.Connections.Lock.RUnlock()
	var c, ok = ep.Connections.List[fd]
	if ok {
		return c.Status, ok
	}
	return -1, ok
}

func (ep *EP) GetConnectionCount() int {
	ep.Connections.Lock.RLock()
	defer ep.Connections.Lock.RUnlock()
	return ep.Connections.Count
}

func (ep *EP) SetConnectionData(fd int, data interface{}) bool {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		conn.Data = data
	}
	return ok
}

func (ep *EP) GetConnectionData(fd int) (interface{}, bool) {
	ep.Connections.Lock.RLock()
	defer ep.Connections.Lock.RUnlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		return conn.Data, ok
	}
	return nil, ok
}

func (ep *EP) UpdateConnectionDataWithFunc(fd int, updateDataFunc UpdateDataFunc) bool {
	ep.Connections.Lock.Lock()
	defer ep.Connections.Lock.Unlock()
	var conn, ok = ep.Connections.List[fd]
	if ok {
		updateDataFunc(conn.Data)
	}
	return ok
}
