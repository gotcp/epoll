package epoll

type ErrorCode int

const (
	ERROR_ACCEPT                ErrorCode = 1
	ERROR_ADD_CONNECTION        ErrorCode = 2
	ERROR_SSL_CONNECTION_CREATE ErrorCode = 3
	ERROR_CLOSE_CONNECTION      ErrorCode = 4
	ERROR_READ                  ErrorCode = 5
	ERROR_SSL_READ              ErrorCode = 6
	ERROR_SSL_WRITE             ErrorCode = 7
	ERROR_EPOLL_WAIT            ErrorCode = 8
	ERROR_STOP                  ErrorCode = 9
	ERROR_POOL_BUFFER           ErrorCode = 10
)
