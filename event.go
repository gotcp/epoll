package epoll

type OnAcceptEvent func(fd int)
type OnCloseEvent func(sequenceId int, fd int)
type OnReceiveEvent func(sequenceId int, fd int, msg []byte, n int)
type OnErrorEvent func(sequenceId int, fd int, code ErrorCode, err error)
