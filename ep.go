package epoll

import (
	"github.com/wuyongjia/pool"
	"github.com/wuyongjia/threadpool"
)

type EP struct {
	Host               string
	Port               int
	Epfd               int
	Fd                 int
	Threads            int
	QueueLength        int
	ReadBuffer         int
	EpollEvents        int
	WaitTimeout        int
	KeepAlive          int
	bufferPool         *pool.Pool               // []byte pool, return *[]byte
	threadPoolSequence *threadpool.PoolSequence // thread pool sequence
	OnAccept           OnAcceptEvent
	OnReceive          OnReceiveEvent
	OnClose            OnCloseEvent
	OnError            OnErrorEvent
}
