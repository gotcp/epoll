package epoll

import (
	"sync"

	"github.com/wuyongjia/pool"
	"github.com/wuyongjia/threadpool"
)

type Conns struct {
	List map[int]*Conn
	Lock *sync.RWMutex
}

type Conn struct {
	Fd         int
	SequenceId int
	Timestamp  int64
}

type EP struct {
	Host               string
	Port               int
	Epfd               int
	Fd                 int
	Connections        *Conns
	Threads            int
	QueueLength        int
	ReadBuffer         int
	WriteBuffer        int
	EpollEvents        int
	WaitTimeout        int
	ReadTimeout        int
	WriteTimeout       int
	KeepAlive          int
	bufferPool         *pool.Pool               // []byte pool, return *[]byte
	threadPoolSequence *threadpool.PoolSequence // thread pool sequence
	OnAccept           OnAcceptEvent
	OnReceive          OnReceiveEvent
	OnClose            OnCloseEvent
	OnError            OnErrorEvent
}
