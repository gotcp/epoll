package epoll

import (
	"golang.org/x/sys/unix"
)

func (ep *EP) listen() {
	var err error
	var i, n int
	var fd int
	var events = make([]unix.EpollEvent, ep.EpollEvents)
	for {
		n, err = unix.EpollWait(ep.Epfd, events, ep.WaitTimeout)
		if err == nil {
			for i = 0; i < n; i++ {
				fd = int(events[i].Fd)
				if fd == ep.Fd {
					ep.InvokeAccept()
				} else if events[i].Events&unix.EPOLLIN != 0 {
					ep.read(fd)
				} else {
					if fd > 0 {
						ep.CloseAction(-1, fd)
					}
				}
			}
		} else {
			if err != unix.EINTR {
				ep.InvokeError(-1, -1, ERROR_EPOLL_WAIT, err)
				break
			}
		}
	}
}
