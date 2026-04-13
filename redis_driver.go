package redis_driver

import (
	"github.com/rseleznev/redis_driver/internal/command"
	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/translator"
)

type Client struct {
	opts *models.Options

	// основной механизм отправки команд и получения результатов
	command.Commander
}

func NewClient(opts *models.Options) (Client, error) {
	initOptions(opts)

	c, err := connection.NewConnection(opts)
	if err != nil {
		return Client{}, err
	}

	t := translator.NewTranslator()

	cmdr := command.NewCommander(c, t, t)

	return Client{
		opts: opts,
		Commander: cmdr,
	}, nil
}