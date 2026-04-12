package connection

import (
	"sync"

	"github.com/rseleznev/redis_driver/internal/models"
)

func NewConnection(opts *models.Options, e epoller) (*Connection, error) {
	return &Connection{}, nil
}

type epoller interface {
	Add(models.PollingUnit) error
	GetError() error
}

type Connection struct {
	socketFd int
	mu sync.Mutex
	sendBuf *models.SendBuf
	recvBuf *models.RecvBuf
	poller epoller
}

func (c *Connection) GetSendBuf() *models.SendBuf {
	return c.sendBuf
}

func (c *Connection) SendAndReceive(*models.SendBuf) (*models.RecvBuf, error) {
	return c.recvBuf, nil
}

func (c *Connection) DrainRecvBuf(*models.RecvBuf) {
	
}