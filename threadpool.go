package epoll

import (
	"github.com/wuyongjia/threadpool"
)

func (s *Server) newThreadPool() *threadpool.Pool {
	var p = threadpool.NewWithFunc(s.NumberOfThreads, s.MaxQueueLength, func(payload interface{}) {
		var req, ok = payload.(*request)
		if ok {
			switch req.Op {
			case OP_RECEIVE:
				s.OnReceive(req.Msg[:req.N], req.Fd)
				bp.Put(&req.Msg)
			case OP_ACCEPT:
				s.OnAccept(req.Fd)
			case OP_CLOSE:
				s.OnClose(req.Fd)
			case OP_ERROR:
				s.OnError(req.Fd, req.ErrCode, req.Err)
			}
			rp.Put(req)
		}
	})
	return p
}
