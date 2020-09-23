package epoll

type OpCode int

const (
	OP_UNKNOW  OpCode = -1
	OP_ACCEPT  OpCode = 1
	OP_RECEIVE OpCode = 2
	OP_CLOSE   OpCode = 3
	OP_ERROR   OpCode = 4
)
