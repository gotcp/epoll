package epoll

type OpCode int

const (
	OP_ACCEPT  OpCode = 1
	OP_ACCEPT1 OpCode = 2
	OP_RECEIVE OpCode = 3
	OP_CLOSE   OpCode = 4
	OP_ERROR   OpCode = 5
)
