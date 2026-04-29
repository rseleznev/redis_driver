package redis_driver

import (
	"context"
	"strconv"
	"time"

	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
)

type Client struct {
	opts *models.Options

	// основной механизм отправки команд и получения результатов
	connection.Connector
}

func NewClient(opts *models.Options) (*Client, error) {
	// инициализируем опции
	initOptions(opts)

	// создаем коннектор
	c, err := connection.NewConnector(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		opts: opts,
		Connector: c,
	}

	// включаем RESP3
	_, err = client.Hello3(context.Background())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Ping проверяет соединение с сервером
func (c *Client) Ping(ctx context.Context) (string, error) {
	args := make([]any, 0, 1)
	args = append(args, "PING")

	var r any
	var err error

	for ctx.Err() == nil {
		r, err = c.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return "", err
		}
		break
	}

	result, ok := r.([]byte)
	if !ok {
		return "", models.ErrDataAssert
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
		r, err = c.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return nil, err
		}
		break
	}

	result, ok := r.(map[string]string)
	if !ok {
		return nil, models.ErrDataAssert
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
		_, err := c.Process(ctx, args) // игнорируем строку "OK"
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return err
		}
		break
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
		r, err = c.Process(ctx, args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess {
				continue
			}
			return nil, err
		}
		break
	}

	result, ok := r.([]byte)
	if !ok {
		return nil, models.ErrDataAssert
	}

	return result, nil
}