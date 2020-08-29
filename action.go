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
			if err = ep.Add(fd); err == nil {
				ep.triggerOnAccept(fd)
			} else {
				ep.triggerOnError(ERROR_ADD_CONNECTION, err)
			}
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
	var sequenceId = tpsequence.GetSequenceId()
	for {
		msg, err = ep.getBytesPoolItem()
		if err != nil {
			ep.triggerOnErrorSequence(sequenceId, ERROR_POOL_BUFFER, err)
			ep.closeActionSequence(sequenceId, fd)
			break
		}
		n, err = unix.Read(fd, *msg)
		if err == nil {
			if n > 0 {
				ep.triggerOnReceiveSequence(sequenceId, fd, msg, n)
			} else {
				bp.Put(msg)
				ep.closeActionSequence(sequenceId, fd)
				break
			}
		} else {
			bp.Put(msg)
			break
		}
	}
}

func (ep *EP) CloseAction(fd int) {
	var err error
	if err = ep.Delete(fd); err == nil {
		ep.triggerOnClose(fd)
	}
}

func (ep *EP) closeActionSequence(sequenceId int, fd int) {
	var err error
	if err = ep.Delete(fd); err == nil {
		ep.triggerOnCloseSequence(sequenceId, fd)
	}
}
