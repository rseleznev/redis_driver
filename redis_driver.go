package redis_driver

import (
	"context"
	"strconv"
	"time"

	"github.com/rseleznev/redis_driver/internal/command"
	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
)

type Client struct {
	opts *models.Options

	// основной механизм отправки команд и получения результатов
	command.Commander
}

func NewClient(opts *models.Options) (*Client, error) {
	// инициализируем опции
	initOptions(opts)

	// создаем соединение
	c, err := connection.NewConnection(opts)
	if err != nil {
		return nil, err
	}

	// создаем главного координатора
	cmdr := command.NewCommander(c)

	// вызываем Hello3()

	return &Client{
		opts: opts,
		Commander: cmdr,
	}, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	args := make([]any, 0, 1)
	args = append(args, "PING")

	return "", nil
}

func (c *Client) Hello(ctx context.Context) (map[string]string, error) {
	args := make([]any, 0, 2)
	args = append(args, "HELLO", "3")

	return nil, nil
}

func (c *Client) Set(ctx context.Context, key string, value any, duration time.Duration) error {
	args := make([]any, 0, 5)
	args = append(args, "SET", key, value)
	if ms := duration.Milliseconds(); ms > 0 {
		msString := strconv.FormatInt(ms, 10)
		args = append(args, "PX", msString)
	}
	
	return nil
}

func (c *Client) Get(ctx context.Context, key string) (any, error) {
	args := make([]any, 0, 2)
	args = append(args, "GET", key)

	return nil, nil
}