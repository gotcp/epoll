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

func (ep *EP) TriggerOnAccept(fd int) {
	if ep.OnAccept != nil {
		ep.threadPoolSequence.Invoke(-1, ep.getRequestItemForAccept(fd))
	}
}

func (ep *EP) TriggerOnClose(fd int) {
	ep.threadPoolSequence.Invoke(-1, ep.getRequestItemForClose(-1, fd))
}

func (ep *EP) TriggerOnError(code ErrorCode, err error) {
	if ep.OnError != nil {
		ep.threadPoolSequence.Invoke(-1, ep.getRequestItemForError(-1, -1, code, err))
	}
}

func (ep *EP) TriggerOnErrorWithFd(fd int, code ErrorCode, err error) {
	if ep.OnError != nil {
		ep.threadPoolSequence.Invoke(-1, ep.getRequestItemForError(-1, fd, code, err))
	}
}

func (ep *EP) TriggerOnReceiveSequence(sequenceId int, fd int, msg *[]byte, n int) {
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForReceive(sequenceId, fd, msg, n))
}

func (ep *EP) TriggerOnCloseSequence(sequenceId int, fd int) {
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForClose(sequenceId, fd))
}

func (ep *EP) TriggerOnErrorSequence(sequenceId int, code ErrorCode, err error) {
	if ep.OnError != nil {
		ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForError(sequenceId, -1, code, err))
	}
}

func (ep *EP) TriggerOnErrorWithFdSequence(sequenceId int, fd int, code ErrorCode, err error) {
	if ep.OnError != nil {
		ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForError(sequenceId, fd, code, err))
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
