package epoll

func (ep *EP) GetSequenceId() int {
	return ep.threadPoolSequence.GetSequenceId()
}
