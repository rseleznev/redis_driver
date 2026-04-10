package command

import (
	"context"
	"time"

	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/translator"
)

// Commander - клиентский интерфейс, который реализует все доступные команды
type Commander interface {
	Ping(context.Context) error
	Hello(context.Context) (map[string]string, error)
	Set(context.Context, string, any, time.Duration) error
	Get(context.Context, string) (any, error)
}

// processor - интерфейс, который реализует отправку команды и получение результата или ошибки
type processor interface {
	sendAndReceive(cmd *command)
}

// commandProcessor обеспечивает отправку команды и получение результата
type commandProcessor struct {
	// абстракция над соединением (в будущем будет пул соединений)
	connector connection.Connector

	// кодировщик в формат RESP3
	enc translator.Encoder

	// декодировщик формата RESP3
	dec translator.Decoder
}

// sendAndReceive осуществляет полный путь команды от сериализации до возврата результата
func (p *commandProcessor) sendAndReceive(cmd *command) {
	// запрашиваем буфер для заполнения
	sBuf := p.connector.GetSendBuf()

	// кодируем в RESP
	data, err := p.enc.Encode(sBuf.Buf, cmd.args)
	if err != nil {
		cmd.resultErrChan <- err

		return
	}
	sBuf.Buf = data

	var rBuf *models.RecvBuf

	// отправляем команду и ждем результат
	rBuf, err = p.connector.SendAndReceive(sBuf)
	if err != nil {
		cmd.resultErrChan <- err

		return
	}

	var result any

	// декодируем в объект Go
	result, err = p.dec.Decode(rBuf.Buf)
	if err != nil {
		cmd.resultErrChan <- err

		return
	}

	// сообщаем, что можно очистить буфер получения
	p.connector.DrainRecvBuf(rBuf)

	// возвращаем успешный результат
	cmd.resultValueChan <- result
}




// ------------------------------------------------
// commandBuilder формирует правильный срез аргументов конкретной команды
type commandBuilder struct {
	// процессор команд (отправка и получение результата/ошибки)
	proc processor
}

// NewClient возвращает клиента, готового выполнять команды
func NewClient(opts models.Options) (Commander, error) {
	c, err := connection.NewC(opts)
	if err != nil {
		return nil, err
	}
	e := translator.NewEncoder()
	d := translator.NewDecoder()
	
	return &commandBuilder{
		proc: &commandProcessor{
			connector: c,
			enc: e,
			dec: d,
		},
	}, nil
}

type command struct {
	args []any
	resultValueChan chan any
	resultErrChan chan error
}

func (b *commandBuilder) Ping(ctx context.Context) error {
	cmd := command{
		args: make([]any, 0, 1),
		resultValueChan: make(chan any),
		resultErrChan: make(chan error),
	}
	cmd.args = append(cmd.args, "PING")
	b.proc.sendAndReceive(&cmd)

	select {
	case err := <-cmd.resultErrChan:
		if err != nil {
			return err
		}

	case <-cmd.resultValueChan:
		return nil

	case <-ctx.Done():
		return ctx.Err()

	}

	return nil
}

func (b *commandBuilder) Hello(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func (b *commandBuilder) Set(context.Context, string, any, time.Duration) error {
	return nil
}

func (b *commandBuilder) Get(context.Context, string) (any, error) {
	return nil, nil
}