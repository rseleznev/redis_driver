package connection

import (
	"sync"

	"github.com/rseleznev/redis_driver/internal/models"
)

func NewConnection(opts models.Options, e epoller) (Connection, error) {
	return Connection{}, nil
}

type Connection struct {
	*connection
}

type epoller interface {
	Add(models.PollingUnit) error
	GetError() error
}

type connection struct {
	socketFd int
	mu sync.Mutex
	sendBuf *models.SendBuf
	recvBuf *models.RecvBuf
	poller epoller
}

func (c *connection) GetSendBuf() *models.SendBuf {
	return c.sendBuf
}

func (c *connection) SendAndReceive(*models.SendBuf) (*models.RecvBuf, error) {
	return c.recvBuf, nil
}

func (c *connection) DrainRecvBuf(*models.RecvBuf) {
	
}