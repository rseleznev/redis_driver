package connection

import (
	"sync"

	"github.com/rseleznev/redis_driver/internal/socket"
	"github.com/rseleznev/redis_driver/internal/transmitter"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/pkg/polling"
)

var (
	epoll *polling.Epoll
	epollErr error
	once sync.Once
)

type Factory struct {}

func (f Factory) NewSocket(opts *models.Options) (socketer, error) {
	socket, err := socket.NewSocket(opts.TCPSendBufLen, opts.TCPReceiveBufLen)
	if err != nil {
		return nil, err
	}

	return socket, nil
}

func (f Factory) NewPoller() (poller, error) {
	// должен создаваться только один инстанс epoll
	once.Do(func() {
		epoll, epollErr = polling.NewEpoll()
	})
	
	if epollErr != nil {
		return nil, epollErr
	}

	return epoll, epollErr
}

func (f Factory) NewSender() sender {
	return transmitter.NewSender()
}

func (f Factory) NewReceiver() receiver {
	return transmitter.NewReceiver()
}