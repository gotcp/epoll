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

func (s *Server) triggerOnAccept(fd int) {
	if s.OnAccept != nil {
		tp.Invoke(s.getRequestItemForAccept(fd))
	}
}

func (s *Server) triggerOnReceive(fd int, msg *[]byte, n int) {
	tp.Invoke(s.getRequestItemForReceive(fd, msg, n))
}

func (s *Server) triggerOnClose(fd int) {
	if s.OnClose != nil {
		tp.Invoke(s.getRequestItemForClose(fd))
	}
}

func (s *Server) triggerOnError(code ErrorCode, err error) {
	if s.OnError != nil {
		tp.Invoke(s.getRequestItemForError(-1, code, err))
	}
}

func (s *Server) triggerOnErrorWithFd(fd int, code ErrorCode, err error) {
	if s.OnError != nil {
		tp.Invoke(s.getRequestItemForError(fd, code, err))
	}
}

func (s *Server) newRequestPool() sync.Pool {
	var p = sync.Pool{
		New: func() interface{} {
			return &request{}
		},
	}
	return p
}

func (s *Server) getRequestItem() *request {
	var item = rp.Get()
	return item.(*request)
}

func (s *Server) getRequestItemForAccept(fd int) *request {
	var item = s.getRequestItem()
	item.Op = OP_ACCEPT
	item.Fd = fd
	return item
}

func (s *Server) getRequestItemForReceive(fd int, msg *[]byte, n int) *request {
	var item = s.getRequestItem()
	item.Op = OP_RECEIVE
	item.Fd = fd
	item.Msg = *msg
	item.N = n
	return item
}

func (s *Server) getRequestItemForClose(fd int) *request {
	var item = s.getRequestItem()
	item.Op = OP_CLOSE
	item.Fd = fd
	return item
}

func (s *Server) getRequestItemForError(fd int, errCode ErrorCode, err error) *request {
	var item = s.getRequestItem()
	item.Op = OP_ERROR
	item.Fd = fd
	item.ErrCode = errCode
	item.Err = err
	return item
}
