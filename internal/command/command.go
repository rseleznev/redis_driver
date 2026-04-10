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

// commandProcessor обеспечивает отправку команды и получение результата
type commandProcessor struct {
	id string
	connector connection.Connector
	enc translator.Encoder
	dec translator.Decoder
}

func (p *commandProcessor) sendAndReceive(cmd *command) {
	// Connector.GetSendBuf

	// Encoder.Encode

	// Connector.SendAndReceive

	// Decoder.Decode

	// Connector.DrainRecvBuf

	// заполняем каналы cmd.resultValueChan и cmd.resultErrChan
}




// ------------------------------------------------
// commandBuilder формирует правильный срез аргументов конкретной команды
type commandBuilder struct {
	processor *commandProcessor
}

// NewClient возвращает клиента, готового выполнять команды
func NewClient(opts models.Options) (Commander, error) {
	conn, err := connection.NewC(opts)
	if err != nil {
		return nil, err
	}
	e := translator.NewEncoder()
	d := translator.NewDecoder()
	
	return &commandBuilder{
		processor: &commandProcessor{
			id: "1111", // надо генерировать или вообще убрать временно
			connector: conn,
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
		resultErrChan: make(chan error),
	}
	cmd.args = append(cmd.args, "PING")
	b.processor.sendAndReceive(&cmd)

	select {
	case err := <-cmd.resultErrChan:
		if err != nil {
			return err
		}

	case <-ctx.Done():
		return ctx.Err()

	}

	return nil
}

func (b *commandBuilder) Hello(ctx context.Context) (map[string]string, error)
func (b *commandBuilder) Set(context.Context, string, any, time.Duration) error
func (b *commandBuilder) Get(context.Context, string) (any, error)