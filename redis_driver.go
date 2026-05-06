package redis_driver

import (
	"context"
	"strconv"
	"time"

	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/options"
)

var (
	ErrTypeAssert = models.ErrTypeAssert
	ErrNoValue = models.ErrNoValue
	ErrOperationRetriesFailed = models.ErrOperationRetriesFailed
)

type Client struct {
	opts *options.Options

	// основной механизм отправки команд и получения результатов
	connector connection.Connector
}

func NewClient(opts *options.Options) (*Client, error) {
	// инициализируем опции
	opts.InitOptions()

	// создаем коннектор
	c, err := connection.NewConnector(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		opts: opts,
		connector: c,
	}

	// включаем RESP3
	if client.opts.ProtoVersion == 3 {
		_, err = client.Hello3(context.Background())
		if err != nil {
			return nil, err
		}	
	}

	return client, nil
}

func (c *Client) GetOptions() *options.Options {
	return c.opts
}

// Ping проверяет соединение с сервером
func (c *Client) Ping(ctx context.Context) (string, error) {
	args := make([]any, 0, 1)
	args = append(args, "PING")

	var r any
	var err error

	for ctx.Err() == nil {
		r, err = c.connector.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return "", err
		}
		break
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	result, ok := r.([]byte)
	if !ok {
		return "", ErrTypeAssert
	}

	return string(result), nil
}

// Hello3 проверяет соединение аналогично Ping, но также возвращает
// дополнительную информацию о клиенте
func (c *Client) Hello3(ctx context.Context) (map[string]string, error) {
	args := make([]any, 0, 2)
	args = append(args, "HELLO", "3")

	var r any
	var err error

	for ctx.Err() == nil {
		r, err = c.connector.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return nil, err
		}
		break
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	result, ok := r.(map[string]string)
	if !ok {
		return nil, ErrTypeAssert
	}

	return result, nil
}

// SetValueForKey устанавливает значение для ключа.
// Если значение уже есть, оно перезаписывается
func (c *Client) SetValueForKey(ctx context.Context, key string, value any, duration time.Duration) error {
	args := make([]any, 0, 5)
	args = append(args, "SET", key, value)
	if ms := duration.Milliseconds(); ms > 0 {
		msString := strconv.FormatInt(ms, 10)
		args = append(args, "PX", msString)
	}

	for ctx.Err() == nil {
		_, err := c.connector.Process(ctx, args) // игнорируем строку "OK"
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return err
		}
		break
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	
	return nil
}

// GetValueByKey получает значение по ключу.
// Всегда возвращает срез байт независимо от того, что было записано
func (c *Client) GetValueByKey(ctx context.Context, key string) ([]byte, error) {
	args := make([]any, 0, 2)
	args = append(args, "GET", key)

	var r any
	var err error

	for ctx.Err() == nil {
		r, err = c.connector.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return nil, err
		}
		break
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	result, ok := r.([]byte)
	if !ok {
		return nil, ErrTypeAssert
	}

	return result, nil
}

func (c *Client) Close() {
	c.connector.Close()
}