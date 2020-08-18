package epoll

import (
	"sync"
)

type request struct {
	Op      OpCode
	Msg     []byte
	N       int
	Fd      int
	ErrCode ErrorCode
	Err     error
}

func (ep *EP) triggerOnAccept(fd int) {
	if ep.OnAccept != nil {
		tp.Invoke(ep.getRequestItemForAccept(fd))
	}
}

func (ep *EP) triggerOnReceive(fd int, msg *[]byte, n int) {
	tp.Invoke(ep.getRequestItemForReceive(fd, msg, n))
}

func (ep *EP) triggerOnClose(fd int) {
	if ep.OnClose != nil {
		tp.Invoke(ep.getRequestItemForClose(fd))
	}
}

func (ep *EP) triggerOnError(code ErrorCode, err error) {
	if ep.OnError != nil {
		tp.Invoke(ep.getRequestItemForError(-1, code, err))
	}
}

func (ep *EP) triggerOnErrorWithFd(fd int, code ErrorCode, err error) {
	if ep.OnError != nil {
		tp.Invoke(ep.getRequestItemForError(fd, code, err))
	}
}

func (ep *EP) newRequestPool() sync.Pool {
	var p = sync.Pool{
		New: func() interface{} {
			return &request{}
		},
	}
	return p
}

func (ep *EP) getRequestItem() *request {
	var item = rp.Get()
	return item.(*request)
}

func (ep *EP) getRequestItemForAccept(fd int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_ACCEPT
	item.Fd = fd
	return item
}

func (ep *EP) getRequestItemForReceive(fd int, msg *[]byte, n int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_RECEIVE
	item.Fd = fd
	item.Msg = *msg
	item.N = n
	return item
}

func (ep *EP) getRequestItemForClose(fd int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_CLOSE
	item.Fd = fd
	return item
}

func (ep *EP) getRequestItemForError(fd int, errCode ErrorCode, err error) *request {
	var item = ep.getRequestItem()
	item.Op = OP_ERROR
	item.Fd = fd
	item.ErrCode = errCode
	item.Err = err
	return item
}
