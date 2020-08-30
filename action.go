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
				ep.TriggerOnAccept(fd)
			} else {
				ep.TriggerOnError(ERROR_ADD_CONNECTION, err)
			}
		} else {
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				ep.TriggerOnError(ERROR_ACCEPT, err)
			}
			break
		}
	}
}

func (ep *EP) readAction(fd int) {
	var err error
	var msg *[]byte
	var n int
	var sequenceId = ep.threadPoolSequence.GetSequenceId()
	for {
		msg, err = ep.GetBufferPoolItem()
		if err != nil {
			ep.TriggerOnErrorSequence(sequenceId, ERROR_POOL_BUFFER, err)
			ep.CloseActionSequence(sequenceId, fd)
			break
		}
		n, err = unix.Read(fd, *msg)
		if err == nil {
			if n > 0 {
				ep.TriggerOnReceiveSequence(sequenceId, fd, msg, n)
			} else {
				ep.bufferPool.Put(msg)
				ep.CloseActionSequence(sequenceId, fd)
				break
			}
		} else {
			ep.bufferPool.Put(msg)
			break
		}
	}
}

func (ep *EP) CloseAction(fd int) {
	var err error
	if err = ep.Delete(fd); err == nil {
		ep.TriggerOnClose(fd)
	}
}

func (ep *EP) CloseActionSequence(sequenceId int, fd int) {
	var err error
	if err = ep.Delete(fd); err == nil {
		ep.TriggerOnCloseSequence(sequenceId, fd)
	}
}
