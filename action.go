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
			if ep.IsSSL {
				var ssl = ep.newSSL(fd)
				if ssl != nil {
					ep.AddConnectionSSL(fd, ssl, sequenceId)
				} else {
					ep.CloseFd(fd)
					if ep.OnError != nil {
						err = errors.New("unable to create SSL connection")
						ep.OnError(fd, ERROR_SSL_CONNECTION_CREATE, err)
					}
					continue
				}
			} else {
				ep.AddConnection(fd, sequenceId)
			}
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
	var sequenceId, ssl = ep.GetConnectionSequenceIdAndSSL(fd)
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
		/**
		if ep.IsSSL {
			if ep.sslHandshakeComplete(ssl) == false {
				if ep.sslHandshake(ssl) == false {
					continue
				}
			}
		}
		*/
		msg, err = ep.GetBufferPoolItem()
		if err != nil {
			ep.InvokeError(sequenceId, fd, ERROR_POOL_BUFFER, err)
			ep.CloseAction(sequenceId, fd)
			break
		}
		if ep.IsSSL {
			n = sslRead(ssl.SSL, *msg, ep.ReadBuffer)
		} else {
			n, err = unix.Read(fd, *msg)
		}
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
