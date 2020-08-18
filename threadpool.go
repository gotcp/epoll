package epoll

import (
	"github.com/wuyongjia/threadpool"
)

func (ep *EP) newThreadPool() *threadpool.Pool {
	var p = threadpool.NewWithFunc(ep.NumberOfThreads, ep.MaxQueueLength, func(payload interface{}) {
		var req, ok = payload.(*request)
		if ok {
			switch req.Op {
			case OP_RECEIVE:
				ep.OnReceive(req.Fd, req.Msg[:req.N], req.N)
				bp.Put(&req.Msg)
			case OP_ACCEPT:
				ep.OnAccept(req.Fd)
			case OP_CLOSE:
				ep.OnClose(req.Fd)
			case OP_ERROR:
				ep.OnError(req.Fd, req.ErrCode, req.Err)
			}
			rp.Put(req)
		}
	})
	return p
}
