package epoll

import (
	"fmt"
	"net"
)

func (ep *EP) tcpAccept(host string, port int) {
	var listen, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	var e net.Error
	var ok bool

	var fd int
	var conn net.Conn

	for {
		conn, err = listen.Accept()
		if err == nil {
			if fd, err = ep.AddConn(conn); err == nil {
				ep.triggerOnAccept1(fd, conn)
			} else {
				ep.triggerOnError(ERROR_ADD_CONNECTION, err)
			}
		} else {
			if e, ok = err.(net.Error); ok && e.Temporary() {
				continue
			}
			ep.triggerOnError(ERROR_ACCEPT1, err)
		}
	}
}
