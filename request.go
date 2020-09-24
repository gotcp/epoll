package epoll

import (
	"errors"
	"fmt"
)

type Request struct {
	Id         uint64
	Op         OpCode
	Fd         int
	Msg        []byte
	N          int
	SequenceId int
	ErrCode    ErrorCode
	Err        error
}

func requestRecycleUpdate(ptr interface{}) {
	var req, ok = ptr.(*Request)
	if ok && req != nil {
		resetRequest(req)
	}
}

func resetRequest(req *Request) {
	req.Msg = nil
	req.Err = nil
}

func (ep *EP) getRequest() *Request {
	var req, err = ep.requestPool.Get()
	if err == nil {
		return req.(*Request)
	}
	return nil
}

func (ep *EP) putRequest(req *Request) {
	ep.requestPool.PutWithId(req, req.Id)
}

func (ep *EP) getRequestItem() *Request {
	return ep.getRequest()
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

func (ep *EP) getRequestItemForAccept(sequenceId int) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_ACCEPT
	req.SequenceId = sequenceId
	return req
}

func (ep *EP) getRequestItemForReceive(sequenceId int, fd int, msg *[]byte, n int) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_RECEIVE
	req.Fd = fd
	req.SequenceId = sequenceId
	req.Msg = *msg
	req.N = n
	return req
}

func (ep *EP) getRequestItemForClose(fd int) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_CLOSE
	req.Fd = fd
	return req
}

func (ep *EP) getRequestItemForError(fd int, errCode ErrorCode, err error) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_ERROR
	req.Fd = fd
	req.ErrCode = errCode
	req.Err = err
	return req
}
