package epoll

import (
	"golang.org/x/sys/unix"
)

func (ep *EP) acceptAction() {
	var err error
	var fd int
	for {
		fd, _, err = unix.Accept(ep.Fd)
		if err == nil {
			if err = ep.Add(fd); err != nil {
				ep.triggerOnError(ERROR_ADD_CONNECTION, err)
				continue
			}
			ep.triggerOnAccept(fd)
		} else {
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				ep.triggerOnError(ERROR_ACCEPT, err)
			}
			break
		}
	}
}

func (ep *EP) readAction(fd int) {
	var err error
	var msg *[]byte
	var n int
	for {
		msg, err = ep.getBytesPoolItem()
		if err != nil {
			ep.CloseAction(fd)
			ep.triggerOnError(ERROR_POOL_BUFFER, err)
			break
		}
		n, err = unix.Read(fd, *msg)
		if err == nil {
			if n > 0 {
				ep.triggerOnReceive(fd, msg, n)
			} else {
				bp.Put(msg)
				ep.CloseAction(fd)
			}
		} else {
			bp.Put(msg)
			break
		}
	}
}

func (ep *EP) CloseAction(fd int) {
	var err error
	if err = ep.Del(fd); err != nil {
		ep.triggerOnErrorWithFd(fd, ERROR_CLOSE_CONNECTION, err)
		return
	}
	ep.triggerOnClose(fd)
}
