package epoll

import (
	"net"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/wuyongjia/pool"
)

const (
	DEFAULT_EPOLL_EVENTS        = 4096
	DEFAULT_EPOLL_READ_TIMEOUT  = 5
	DEFAULT_EPOLL_WRITE_TIMEOUT = 5
	DEFAULT_POOL_MULTIPLE       = 12
)

func New(readBuffer int, threads int, queueLength int) (*EP, error) {
	var epfd, err = unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	var conns = &Conns{
		List:  make(map[int]*Conn),
		Count: 0,
		Lock:  &sync.RWMutex{},
	}
	var ep = &EP{
		Epfd:         epfd,
		Fd:           -9,
		Connections:  conns,
		SSLCtx:       nil,
		IsSSL:        false,
		ReadBuffer:   readBuffer,
		WriteBuffer:  readBuffer,
		WaitTimeout:  -1,
		ReadTimeout:  DEFAULT_EPOLL_READ_TIMEOUT,
		WriteTimeout: DEFAULT_EPOLL_WRITE_TIMEOUT,
		Threads:      threads,
		QueueLength:  queueLength,
		KeepAlive:    0,
		ReuseAddr:    1,
		ReusePort:    1,
		EpollEvents:  DEFAULT_EPOLL_EVENTS,
		OnAccept:     nil,
		OnReceive:    nil,
		OnClose:      nil,
		OnError:      nil,
	}
	ep.bufferPool = ep.newBufferPool(readBuffer, threads*DEFAULT_POOL_MULTIPLE)
	ep.connPool = ep.newConnPool(threads * DEFAULT_POOL_MULTIPLE)
	ep.requestPool = ep.newRequestPool(threads * DEFAULT_POOL_MULTIPLE)
	ep.threadPoolSequence = ep.newThreadPoolSequence()
	return ep, nil
}

func (ep *EP) newBufferPool(readBuffer int, length int) *pool.Pool {
	return pool.New(length, func() interface{} {
		var b = make([]byte, readBuffer)
		return &b
	})
}

func (ep *EP) newConnPool(length int) *pool.Pool {
	return pool.NewWithId(length, func(id uint64) interface{} {
		return &Conn{Id: id}
	})
}

func (ep *EP) newRequestPool(length int) *pool.Pool {
	return pool.NewWithId(length, func(id uint64) interface{} {
		return &Request{Id: id}
	})
}

func (ep *EP) SetWaitTimeout(n int) {
	ep.WaitTimeout = n
}

func (ep *EP) SetReadTimeout(n int) {
	ep.ReadTimeout = n
}

func (ep *EP) SetWriteTimeout(n int) {
	ep.WriteTimeout = n
}

func (ep *EP) SetWriteBuffer(n int) {
	ep.WriteBuffer = n
}

func (ep *EP) SetKeepAlive(n int) {
	ep.KeepAlive = n
}

func (ep *EP) SetReuseAddr(n int) {
	ep.ReuseAddr = n
}

func (ep *EP) SetReusePort(n int) {
	ep.ReusePort = n
}

func (ep *EP) SetEpollEvents(n int) {
	ep.EpollEvents = n
}

// pure EPOLL
func (ep *EP) Start(host string, port int) {
	ep.Host = host
	ep.Port = port
	var err error
	if err = ep.InitEpoll(ep.Host, ep.Port); err != nil {
		panic(err)
	}
	ep.listen()
}

func (ep *EP) StartSSL(host string, port int, certFile string, keyFile string) {
	ep.IsSSL = true
	ep.Host = host
	ep.Port = port
	var err error
	if err = ep.InitEpoll(ep.Host, ep.Port); err != nil {
		panic(err)
	}
	ep.SSLCtx = newSSLCtx(certFile, keyFile)
	ep.sslPool = ep.newSSLPool(ep.Threads * DEFAULT_POOL_MULTIPLE)

	ep.bufferPool.SetMode(pool.MODE_QUEUE)
	ep.connPool.SetMode(pool.MODE_QUEUE)
	ep.requestPool.SetMode(pool.MODE_QUEUE)
	ep.sslPool.SetMode(pool.MODE_QUEUE)

	cMallocTrimLoop()
	ep.listen()
}

// pure EPOLL, only listening, needs to use ep.Add(fd)
func (ep *EP) Listen() {
	ep.listen()
}

func (ep *EP) InitEpoll(host string, port int) error {
	var err error

	if ep.Fd, err = unix.Socket(unix.AF_INET, unix.O_NONBLOCK|unix.SOCK_STREAM, 0); err != nil {
		return err
	}

	if err = unix.SetNonblock(ep.Fd, true); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, 1); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.IPPROTO_TCP, unix.TCP_QUICKACK, 1); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, ep.ReuseAddr); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, ep.ReusePort); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	var addr = unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(host).To4())

	if err = unix.Bind(ep.Fd, &addr); err != nil {
		unix.Close(ep.Fd)
		return err
	}
	if err = unix.Listen(ep.Fd, unix.SOMAXCONN); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	var event unix.EpollEvent
	event.Events = unix.EPOLLIN | unix.EPOLLET
	event.Fd = int32(ep.Fd)

	if err = unix.EpollCtl(ep.Epfd, unix.EPOLL_CTL_ADD, ep.Fd, &event); err != nil {
		return err
	}

	return nil
}

func (ep *EP) Stop() error {
	ep.CloseAll()
	if ep.Fd >= 0 {
		if ep.Delete(ep.Fd) == nil {
			ep.Close(ep.Fd)
		}
	}
	if ep.Epfd >= 0 {
		unix.Close(ep.Epfd)
	}
	if ep.SSLCtx != nil {
		ep.freeSSLCtx()
	}
	ep.threadPoolSequence.Close()
	return nil
}
