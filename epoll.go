package epoll

import (
	"net"

	"golang.org/x/sys/unix"

	"github.com/wuyongjia/bytespool"
	"github.com/wuyongjia/threadpool"
)

const (
	DEFAULT_EPOLL_EVENTS = 2048
)

var bp *bytespool.Pool  // []byte pool, return *[]byte
var tp *threadpool.Pool // thread pool

func New(readBuffer int, numberOfThreads int, maxQueueLength int) (*EP, error) {
	var epfd, err = unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	var ep = &EP{
		Epfd:            epfd,
		Fd:              -9,
		ReadBuffer:      readBuffer,
		NumberOfThreads: numberOfThreads,
		MaxQueueLength:  maxQueueLength,
		Timeout:         -1,
		KeepAlive:       0,
		MaxEpollEvents:  DEFAULT_EPOLL_EVENTS,
		OnAccept:        nil,
		OnReceive:       nil,
		OnClose:         nil,
		OnError:         nil,
	}

	bp = bytespool.New(readBuffer)
	tp = ep.newThreadPool()

	return ep, nil
}

func (ep *EP) SetTimeout(n int) {
	ep.Timeout = n
}

func (ep *EP) SetKeepAlive(n int) {
	ep.KeepAlive = n
}

func (ep *EP) SetMaxEpollEvents(n int) {
	ep.MaxEpollEvents = n
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
	var err error
	if ep.Fd > 0 {
		if err = ep.Del(ep.Fd); err != nil {
			return err
		}
	}
	if err = unix.Close(ep.Epfd); err != nil {
		return err
	}
	return nil
}
