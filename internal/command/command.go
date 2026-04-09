package command

import (
	"time"

	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Commander клиентский интерфейс, который реализует все доступные команды
type Commander interface {
	Ping() error
	Hello() (map[string]string, error)
	Set(string, any, time.Duration) error
	Get(string) (any, error)
}

// commandProcessor обеспечивает отправку команды и получение результата
type commandProcessor struct {
	id string
	connector connection.Connector
	// интерфейс Encoder
	// интерфейс Decoder
}

func (p *commandProcessor) sendAndReceive(cmd *command) {
	// запрашиваем буфер отправки у Connector

	// Encoder.Encode

	// Connector.SendAndReceive

	// Decoder.Decode

	// заполняем cmd.resultValue
}




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
	
	return &commandBuilder{
		processor: &commandProcessor{
			id: "1111", // надо генерировать или вообще убрать временно
			connector: conn,
		},
	}, nil
}

type command struct {
	args []any
	resultValue any
	resultErr error
}

func (b *commandBuilder) Ping() error {
	cmd := command{
		args: make([]any, 0, 1),
	}
	cmd.args = append(cmd.args, "PING")
	b.processor.sendAndReceive(&cmd)

	if err := cmd.resultErr; err != nil {
		return err
	}

	return nil
}

func (b *commandBuilder) Hello() (map[string]string, error)
func (b *commandBuilder) Set(string, any, time.Duration) error
func (b *commandBuilder) Get(string) (any, error)