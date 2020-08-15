package epoll

import (
	"net"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/wuyongjia/bytespool"
	"github.com/wuyongjia/threadpool"
)

const (
	MAX_LENGTH = 2048
)

var bp *bytespool.Pool  // []byte pool, return *[]byte
var tp *threadpool.Pool // thread pool
var rp sync.Pool        // request pool, return *request

func New(host string, port int, readBuffer int, numberOfThreads int, maxQueueLength int) (*Server, error) {
	var epfd, err = unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Server{
		Host:            host,
		Port:            port,
		Epfd:            epfd,
		ReadBuffer:      readBuffer,
		NumberOfThreads: numberOfThreads,
		MaxQueueLength:  maxQueueLength,
		Timeout:         -1,
		KeepAlive:       0,
		MaxEpollEvents:  MAX_LENGTH,
		OnAccept:        nil,
		OnReceive:       nil,
		OnClose:         nil,
		OnError:         nil,
	}, nil
}

func (s *Server) SetTimeout(n int) {
	s.Timeout = n
}

func (s *Server) SetKeepAlive(n int) {
	s.KeepAlive = n
}

func (s *Server) SetMaxEpollEvents(n int) {
	s.MaxEpollEvents = n
}

func (s *Server) Start() {
	var err error
	if err = s.initEpoll(s.Host, s.Port); err != nil {
		panic(err)
	}

	bp = bytespool.New(s.ReadBuffer)
	tp = s.newThreadPool()
	rp = s.newRequestPool()

	s.listen()
}

func (s *Server) initEpoll(host string, port int) error {
	var err error

	if s.Fd, err = unix.Socket(unix.AF_INET, unix.O_NONBLOCK|unix.SOCK_STREAM, 0); err != nil {
		return err
	}
	if err = unix.SetNonblock(s.Fd, true); err != nil {
		unix.Close(s.Fd)
		return err
	}
	if err = unix.SetsockoptInt(s.Fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, 1); err != nil {
		unix.Close(s.Fd)
		return err
	}
	if err = unix.SetsockoptInt(s.Fd, unix.IPPROTO_TCP, unix.TCP_QUICKACK, 1); err != nil {
		unix.Close(s.Fd)
		return err
	}

	var addr = unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(host).To4())

	if err = unix.Bind(s.Fd, &addr); err != nil {
		unix.Close(s.Fd)
		return err
	}
	if err = unix.Listen(s.Fd, unix.SOMAXCONN); err != nil {
		unix.Close(s.Fd)
		return err
	}

	var event unix.EpollEvent
	event.Events = unix.EPOLLIN | unix.EPOLLET
	event.Fd = int32(s.Fd)

	if err = unix.EpollCtl(s.Epfd, unix.EPOLL_CTL_ADD, s.Fd, &event); err != nil {
		return err
	}

	return nil
}

func (s *Server) listen() {
	var err error
	var i, n int
	var fd int

	var events = make([]unix.EpollEvent, s.MaxEpollEvents)

	for {
		n, err = unix.EpollWait(s.Epfd, events, s.Timeout)
		if err == nil {
			for i = 0; i < n; i++ {
				fd = int(events[i].Fd)
				if fd == s.Fd {
					s.accept()
				} else if events[i].Events&unix.EPOLLIN != 0 {
					s.read(fd)
				} else if events[i].Events&(unix.EPOLLERR|unix.EPOLLHUP|unix.EPOLLRDHUP) != 0 {
					s.Close(fd)
				} else {
					s.triggerOnError(ERROR_OTHER_EVENTS, nil)
				}
			}
		} else {
			if err != unix.EINTR {
				s.triggerOnError(ERROR_EPOLL_WAIT, err)
				break
			}
		}
	}
}

func (s *Server) read(fd int) {
	var err error
	var msg *[]byte
	var n int
	for {
		msg = bp.Get()
		n, err = unix.Read(fd, *msg)
		if err == nil {
			if n > 0 {
				tp.Invoke(s.getRequestItemForReceive(fd, msg, n))
			} else {
				bp.Put(msg)
				s.Close(fd)
			}
		} else {
			bp.Put(msg)
			// if !(err == unix.EAGAIN || err == unix.EWOULDBLOCK) {s.triggerOnError(ERROR_READ, err)}
			break
		}
	}
}

func (s *Server) accept() {
	var err error
	var fd int
	for {
		fd, _, err = unix.Accept(s.Fd)
		if err != nil {
			if (err == unix.EAGAIN) || (err == unix.EWOULDBLOCK) {
				break
			} else {
				s.triggerOnError(ERROR_ACCEPT, err)
				break
			}
		}

		err = unix.EpollCtl(s.Epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLET, Fd: int32(fd)})
		if err != nil {
			s.triggerOnError(ERROR_ADD_CONNECTION, err)
			break
		}

		unix.SetNonblock(fd, true)
		unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, s.KeepAlive)
		unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, s.ReadBuffer)

		s.triggerOnAccept(fd)
	}
}

func (s *Server) Close(fd int) {
	var err error
	if err = unix.EpollCtl(s.Epfd, unix.EPOLL_CTL_DEL, fd, nil); err != nil {
		s.triggerOnErrorWithFd(fd, ERROR_CLOSE_CONNECTION, err)
		return
	}
	if err = unix.Close(fd); err != nil {
		s.triggerOnErrorWithFd(fd, ERROR_CLOSE_CONNECTION, err)
		return
	}
	s.triggerOnClose(fd)
}

func (s *Server) Stop() {
	s.Close(s.Fd)
	var err error
	if err = unix.Close(s.Epfd); err != nil {
		s.triggerOnError(ERROR_STOP, err)
	}
}

func Write(fd int, msg []byte) error {
	var _, err = unix.Write(fd, msg)
	return err
}
