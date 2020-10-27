package epoll

type OpCode int

const (
	OP_UNKNOW   OpCode = -1
	OP_ACCEPT   OpCode = 1
	OP_RECEIVE  OpCode = 2
	OP_EPOLLOUT OpCode = 3
	OP_CLOSE    OpCode = 4
	OP_ERROR    OpCode = 5
)
