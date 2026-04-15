package redis_driver

import (
	"github.com/rseleznev/redis_driver/internal/command"
	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
)

type Client struct {
	opts *models.Options

	// основной механизм отправки команд и получения результатов
	command.Commander
}

func NewClient(opts *models.Options) (Client, error) {
	// инициализируем опции
	initOptions(opts)

	// создаем соединение
	c, err := connection.NewConnection(opts)
	if err != nil {
		return Client{}, err
	}

	// создаем главного координатора
	cmdr := command.NewCommander(c)

	// вызываем Hello3()

	return Client{
		opts: opts,
		Commander: cmdr,
	}, nil
}