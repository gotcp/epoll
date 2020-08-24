package epoll

type ErrorCode int

const (
	ERROR_ACCEPT           ErrorCode = 1
	ERROR_ADD_CONNECTION   ErrorCode = 2
	ERROR_CLOSE_CONNECTION ErrorCode = 3
	ERROR_EPOLL_WAIT       ErrorCode = 4
	ERROR_STOP             ErrorCode = 5
	ERROR_POOL_BUFFER      ErrorCode = 6
)
