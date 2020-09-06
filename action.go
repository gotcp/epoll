package epoll

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

func (ep *EP) accept(sequenceId int) {
	var err error
	var fd int
	for {
		fd, _, err = unix.Accept(ep.Fd)
		if err == nil {
			ep.AddConnection(fd, sequenceId)
			if err = ep.Add(fd); err == nil {
				if ep.OnAccept != nil {
					ep.OnAccept(fd)
				}
			} else {
				ep.DeleteConnection(fd)
				if ep.OnError != nil {
					ep.OnError(fd, ERROR_ADD_CONNECTION, err)
				}
			}
		} else {
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				if ep.OnError != nil {
					ep.OnError(fd, ERROR_ACCEPT, err)
				}
			}
			break
		}
	}
}

func (ep *EP) read(fd int) {
	var err error
	var sequenceId = ep.GetConnectionSequenceId(fd)
	if sequenceId < 0 {
		if ep.OnError != nil {
			err = errors.New(fmt.Sprintf("%d not found in the list", fd))
			ep.InvokeError(-1, fd, ERROR_READ, err)
		}
		return
	}
	var msg *[]byte
	var n int
	for {
		msg, err = ep.GetBufferPoolItem()
		if err != nil {
			ep.InvokeError(sequenceId, fd, ERROR_POOL_BUFFER, err)
			ep.CloseAction(sequenceId, fd)
			break
		}
		n, err = unix.Read(fd, *msg)
		if err == nil {
			if n > 0 {
				ep.InvokeReceive(sequenceId, fd, msg, n)
			} else {
				ep.PutBufferPoolItem(msg)
				ep.CloseAction(sequenceId, fd)
				break
			}
		} else {
			ep.PutBufferPoolItem(msg)
			break
		}
	}
}

func (ep *EP) CloseAction(sequenceId int, fd int) error {
	var err error
	if err = ep.Delete(fd); err == nil {
		ep.InvokeClose(sequenceId, fd)
	}
	return err
}
