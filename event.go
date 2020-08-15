package epoll

type OnAcceptEvent func(fd int)
type OnCloseEvent func(fd int)
type OnReceiveEvent func(msg []byte, fd int)
type OnErrorEvent func(fd int, code ErrorCode, err error)
