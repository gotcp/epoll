package epoll

import (
	"errors"
	"fmt"
	"sync"
)

type request struct {
	Op         OpCode
	Fd         int
	Msg        []byte
	N          int
	SequenceId int
	ErrCode    ErrorCode
	Err        error
}

func (ep *EP) InvokeAccept() {
	var sequenceId = ep.GetSequenceId()
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForAccept(sequenceId))
}

func (ep *EP) InvokeReceive(sequenceId int, fd int, msg *[]byte, n int) {
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForReceive(sequenceId, fd, msg, n))
}

func (ep *EP) InvokeClose(sequenceId int, fd int) {
	if sequenceId < 0 {
		sequenceId = ep.GetConnectionSequenceId(fd)
		if sequenceId < 0 {
			if ep.OnError != nil {
				var err = errors.New(fmt.Sprintf("%d not found in the list", fd))
				ep.InvokeError(-1, fd, ERROR_CLOSE_CONNECTION, err)
			}
			return
		}
	}
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForClose(fd))
}

func (ep *EP) InvokeError(sequenceId int, fd int, code ErrorCode, err error) {
	ep.threadPoolSequence.Invoke(sequenceId, ep.getRequestItemForError(fd, code, err))
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

func (ep *EP) getRequestItemForAccept(sequenceId int) *request {
	var item = ep.getRequestItem()
	item.Op = OP_ACCEPT
	item.SequenceId = sequenceId
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
