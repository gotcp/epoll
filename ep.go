package epoll

type EP struct {
	Host            string
	Port            int
	Epfd            int
	Fd              int
	NumberOfThreads int
	MaxQueueLength  int
	ReadBuffer      int
	MaxEpollEvents  int
	Timeout         int
	KeepAlive       int
	OnAccept        OnAcceptEvent
	OnReceive       OnReceiveEvent
	OnClose         OnCloseEvent
	OnError         OnErrorEvent
}
