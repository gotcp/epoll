package epoll

type OnAcceptEvent func(fd int)
type OnCloseEvent func(fd int)
type OnReceiveEvent func(fd int, msg []byte, n int)
type OnEpollOutEvent func(fd int)
type OnErrorEvent func(fd int, code ErrorCode, err error)
