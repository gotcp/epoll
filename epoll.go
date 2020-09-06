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
)

func New(readBuffer int, threads int, queueLength int) (*EP, error) {
	var epfd, err = unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	var conns = &Conns{
		List: make(map[int]*Conn),
		Lock: &sync.RWMutex{},
	}
	var ep = &EP{
		Epfd:         epfd,
		Fd:           -9,
		Connections:  conns,
		ReadBuffer:   readBuffer,
		WriteBuffer:  readBuffer,
		WaitTimeout:  -1,
		ReadTimeout:  DEFAULT_EPOLL_READ_TIMEOUT,
		WriteTimeout: DEFAULT_EPOLL_WRITE_TIMEOUT,
		Threads:      threads,
		QueueLength:  queueLength,
		KeepAlive:    0,
		EpollEvents:  DEFAULT_EPOLL_EVENTS,
		OnAccept:     nil,
		OnReceive:    nil,
		OnClose:      nil,
		OnError:      nil,
	}
	ep.bufferPool = pool.New(20*threads, func() interface{} {
		var buf = make([]byte, readBuffer)
		return &buf
	})
	ep.threadPoolSequence = ep.newThreadPoolSequence()
	return ep, nil
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
	ep.threadPoolSequence.Close()
	return nil
}
