package redis_driver

import (
	"sync"

	"github.com/rseleznev/redis_driver/internal/command"
	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/translator"
	"github.com/rseleznev/redis_driver/pkg/polling"
)

var (
	epoll polling.Epoll
	epollErr error
	once sync.Once
)

type Client struct {
	command.Commander
}

func NewClient(opts models.Options) (Client, error) {
	initOptions(&opts)
	
	once.Do(func() {
		epoll, epollErr = polling.NewEpoll()
	})
	if epollErr != nil {
		return Client{}, epollErr
	}

	c, err := connection.NewConnection(opts, epoll)
	if err != nil {
		return Client{}, err
	}

	t := translator.NewTranslator()

	cmdr := command.NewCommander(c, t, t)

	return Client{
		Commander: cmdr,
	}, nil
}