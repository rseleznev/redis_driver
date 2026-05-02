package models

type PollingUnit struct {
	SocketFd int
	EventType string // connect, income, outcome
	ResultChan chan error // канал, чтобы вызывающий поток заблокировался на чтении
}

type PollingResult struct {
	Err error
}

type SendBuf struct {
	WritePos int
	SentBytes int
	Buf []byte
}

type RecvBuf struct {
	WritePos int
	Buf []byte
}