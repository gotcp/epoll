package epoll

import (
	"sync"
)

type request struct {
	Op         OpCode
	Msg        []byte
	N          int
	Fd         int
	SequenceId int
	ErrCode    ErrorCode
	Err        error
}

func (ep *EP) triggerOnAccept(fd int) {
	if ep.OnAccept != nil {
		tpsequence.Invoke(-1, ep.getRequestItemForAccept(fd))
	}
}

func (ep *EP) triggerOnClose(fd int) {
	tpsequence.Invoke(-1, ep.getRequestItemForClose(-1, fd))
}

func (ep *EP) triggerOnError(code ErrorCode, err error) {
	if ep.OnError != nil {
		tpsequence.Invoke(-1, ep.getRequestItemForError(-1, -1, code, err))
	}
}

func (ep *EP) triggerOnErrorWithFd(fd int, code ErrorCode, err error) {
	if ep.OnError != nil {
		tpsequence.Invoke(-1, ep.getRequestItemForError(-1, fd, code, err))
	}
}

func (ep *EP) triggerOnReceiveSequence(sequenceId int, fd int, msg *[]byte, n int) {
	tpsequence.Invoke(sequenceId, ep.getRequestItemForReceive(sequenceId, fd, msg, n))
}

func (ep *EP) triggerOnCloseSequence(sequenceId int, fd int) {
	tpsequence.Invoke(sequenceId, ep.getRequestItemForClose(sequenceId, fd))
}

func (ep *EP) triggerOnErrorSequence(sequenceId int, code ErrorCode, err error) {
	if ep.OnError != nil {
		tpsequence.Invoke(sequenceId, ep.getRequestItemForError(sequenceId, -1, code, err))
	}
}

func (ep *EP) triggerOnErrorWithFdSequence(sequenceId int, fd int, code ErrorCode, err error) {
	if ep.OnError != nil {
		tpsequence.Invoke(sequenceId, ep.getRequestItemForError(sequenceId, fd, code, err))
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
	return &request{SequenceId: -1}
}

func (ep *EP) getRequestItemForAccept(fd int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_ACCEPT
	item.Fd = fd
	return item
}

func (ep *EP) getRequestItemForReceive(sequenceId int, fd int, msg *[]byte, n int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_RECEIVE
	item.Fd = fd
	item.SequenceId = sequenceId
	item.Msg = *msg
	item.N = n
	return item
}

func (ep *EP) getRequestItemForClose(sequenceId int, fd int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_CLOSE
	item.Fd = fd
	item.SequenceId = sequenceId
	return item
}

func (ep *EP) getRequestItemForError(sequenceId int, fd int, errCode ErrorCode, err error) *request {
	var item = ep.getRequestItem()
	item.Op = OP_ERROR
	item.Fd = fd
	item.SequenceId = sequenceId
	item.ErrCode = errCode
	item.Err = err
	return item
}
