package epoll

import (
	"github.com/wuyongjia/threadpool"
)

func (ep *EP) newThreadPoolSequence() *threadpool.PoolSequence {
	var p = threadpool.NewSequenceWithFunc(ep.Threads, ep.QueueLength, func(payload interface{}) {
		var req, ok = payload.(*request)
		if ok {
			switch req.Op {
			case OP_ACCEPT:
				ep.OnAccept(req.Fd)
			case OP_RECEIVE:
				ep.OnReceive(req.Fd, req.Msg[:req.N], req.N)
				ep.PutBufferPoolItem(&req.Msg)
			case OP_CLOSE:
				if ep.OnClose != nil {
					ep.OnClose(req.Fd)
				}
				ep.Close(req.Fd)
			case OP_ERROR:
				ep.OnError(req.Fd, req.ErrCode, req.Err)
			}
		}
	})
	return p
}
