package connection

import (
	"sync"

	"github.com/rseleznev/redis_driver/internal/models"
)

type Connector interface {
	GetSendBuf() *models.SendBuf
	SendAndReceive(*models.SendBuf) (*models.RecvBuf, error)
	DrainRecvBuf(*models.RecvBuf)
}

func NewC(opts models.Options) (Connector, error) {
	return &connection{}, nil
}

type connection struct {
	socketFd int
	mu sync.Mutex
	sendBuf *models.SendBuf
	recvBuf *models.RecvBuf
}

func (c *connection) GetSendBuf() *models.SendBuf {
	return c.sendBuf
}

func (c *connection) SendAndReceive(*models.SendBuf) (*models.RecvBuf, error) {
	return c.recvBuf, nil
}

func (c *connection) DrainRecvBuf(*models.RecvBuf) {
	
}