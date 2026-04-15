package command

import (
	"context"
	"strconv"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
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

type connector interface {
	GetSendBuf() (*models.SendBuf, error)
	CancelProcessing()
	SendAndReceive(*models.SendBuf) (*models.RecvBuf, error)
	DrainRecvBuf(*models.RecvBuf)
}

type encoder interface {
	Encode([]byte, []any) ([]byte, error)
}

type decoder interface {
	Decode([]byte) (any, error)
}

// commandProcessor обеспечивает отправку команды и получение результата
type commandProcessor struct {
	// абстракция над соединением (в будущем будет пул соединений)
	connector connector

	// кодировщик в формат RESP3
	enc encoder

	// декодировщик формата RESP3
	dec decoder
}

// sendAndReceive осуществляет полный путь команды от сериализации до возврата результата
func (p *commandProcessor) sendAndReceive(cmd *command) {
	if !cmd.isWaiting() {
		return
	}
	var sBuf *models.SendBuf
	var err error
	
	for {
		// запрашиваем буфер для заполнения
		sBuf, err = p.connector.GetSendBuf()
		if err != nil { // пока здесь может быть только одна ошибка
			if err == models.ErrConnectionCmdInProcess && cmd.isWaiting() {
				continue
			}

			select {
			case cmd.resultErrChan <- err:

			case <-cmd.done():

			}
			return
		}
		break
	}

	if !cmd.isWaiting() {
		p.connector.CancelProcessing()
		return
	}

	// кодируем в RESP
	data, err := p.enc.Encode(sBuf.Buf, cmd.args)
	if err != nil {
		select {
		case cmd.resultErrChan <- err:

		case <-cmd.done():
			p.connector.CancelProcessing()

		}
		return
	}
	if !cmd.isWaiting() {
		p.connector.CancelProcessing()
		return
	}
	sBuf.Buf = data

	var rBuf *models.RecvBuf

	// отправляем команду и ждем результат
	rBuf, err = p.connector.SendAndReceive(sBuf)
	if err != nil {
		select {
		case cmd.resultErrChan <- err:

		case <-cmd.done():
			p.connector.CancelProcessing()

		}
		return
	}
	if !cmd.isWaiting() {
		p.connector.CancelProcessing()
		return
	}

	var result any

	// декодируем в объект Go
	result, err = p.dec.Decode(rBuf.Buf)
	if err != nil {
		select {
		case cmd.resultErrChan <- err:

		case <-cmd.done():
			p.connector.CancelProcessing()

		}
		return
	}

	// сообщаем, что можно очистить буфер получения
	p.connector.DrainRecvBuf(rBuf)

	// возвращаем успешный результат
	select {
	case cmd.resultValueChan <- result:

	case <-cmd.done():

	}
	
}




// ------------------------------------------------

// commandBuilder формирует правильный срез аргументов конкретной команды
type commandBuilder struct {
	// процессор команд (отправка и получение результата/ошибки)
	proc processor
}

// NewCommander возвращает клиента, готового выполнять команды
func NewCommander(c connector, e encoder, d decoder) Commander {
	return &commandBuilder{
		proc: &commandProcessor{
			connector: c,
			enc: e,
			dec: d,
		},
	}
}

type command struct {
	args []any
	resultValueChan chan any
	resultErrChan chan error

	timeout chan struct{}
	waiting bool
}

func (c *command) isWaiting() bool {
	return c.waiting
}

func (c *command) stopWaiting() {
	c.waiting = false
	close(c.timeout)
}

func (c *command) done() <-chan struct{} {
	return c.timeout
}

func (b *commandBuilder) Ping(ctx context.Context) error {
	cmd := &command{
		args: make([]any, 0, 1),
		resultValueChan: make(chan any),
		resultErrChan: make(chan error),
		timeout: make(chan struct{}),
		waiting: true,
	}
	cmd.args = append(cmd.args, "PING")
	go b.proc.sendAndReceive(cmd)

	select {
	case err := <-cmd.resultErrChan:
		return err

	case <-cmd.resultValueChan:
		return nil

	case <-ctx.Done():
		cmd.stopWaiting()
		return ctx.Err()

	}
}

func (b *commandBuilder) Hello(ctx context.Context) (map[string]string, error) {
	cmd := &command{
		args: make([]any, 0, 2),
		resultValueChan: make(chan any),
		resultErrChan: make(chan error),
		timeout: make(chan struct{}),
		waiting: true,
	}
	cmd.args = append(cmd.args, "HELLO", "3")
	go b.proc.sendAndReceive(cmd)

	select {
	case err := <-cmd.resultErrChan:
		return nil, err

	case r := <-cmd.resultValueChan:
		res, ok := r.(map[string]string)
		if !ok {
			return nil, models.ErrDataAssert
		}

		return res, nil

	case <-ctx.Done():
		cmd.stopWaiting()
		return nil, ctx.Err()

	}
}

func (b *commandBuilder) Set(ctx context.Context, key string, value any, duration time.Duration) error {
	cmd := &command{
		args: make([]any, 0, 5),
		resultValueChan: make(chan any),
		resultErrChan: make(chan error),
		timeout: make(chan struct{}),
		waiting: true,
	}
	cmd.args = append(cmd.args, "SET", key, value)
	if ms := duration.Milliseconds(); ms > 0 {
		msString := strconv.FormatInt(ms, 10)
		cmd.args = append(cmd.args, "PX", msString)
	}
	go b.proc.sendAndReceive(cmd)

	select {
	case err := <-cmd.resultErrChan:
		return err

	case <-cmd.resultValueChan:
		return nil

	case <-ctx.Done():
		cmd.stopWaiting()
		return ctx.Err()

	}
}

func (b *commandBuilder) Get(ctx context.Context, key string) (any, error) {
	cmd := &command{
		args: make([]any, 0, 2),
		resultValueChan: make(chan any),
		resultErrChan: make(chan error),
		timeout: make(chan struct{}),
		waiting: true,
	}
	cmd.args = append(cmd.args, "GET", key)
	go b.proc.sendAndReceive(cmd)

	select {
	case err := <-cmd.resultErrChan:
		return nil, err

	case r := <-cmd.resultValueChan:
		return r, nil

	case <-ctx.Done():
		cmd.stopWaiting()
		return nil, ctx.Err()

	}
}