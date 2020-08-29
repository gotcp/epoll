package epoll

type EP struct {
	Host        string
	Port        int
	Epfd        int
	Fd          int
	Threads     int
	QueueLength int
	ReadBuffer  int
	EpollEvents int
	WaitTimeout int
	KeepAlive   int
	OnAccept    OnAcceptEvent
	OnReceive   OnReceiveEvent
	OnClose     OnCloseEvent
	OnError     OnErrorEvent
}
