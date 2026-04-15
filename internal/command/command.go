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
	process(cmd *command)
}

type connector interface {
	Process([]any) (any, error)
	CancelProcessing()
	Done()
}

// commandProcessor обеспечивает отправку команды и получение результата
type commandProcessor struct {
	// абстракция над соединением (в будущем будет пул соединений)
	connector connector
}

// process осуществляет полный путь команды от сериализации до возврата результата
func (p *commandProcessor) process(cmd *command) {
	if !cmd.isWaiting() {
		return
	}

	var result any
	var err error

	for {
		result, err = p.connector.Process(cmd.args)
		if err != nil {
			if err == models.ErrConnectionCmdInProcess && cmd.isWaiting() {
				continue
			}
			if !cmd.isWaiting() {
				p.connector.CancelProcessing()
				return
			}

			select {
			case cmd.resultErrChan <- err:

			case <-cmd.done():

			}
			return

		}
		break
	}
	
	// скопировать результат?

	// возвращаем успешный результат
	select {
	case cmd.resultValueChan <- result:

	case <-cmd.done():

	}
	
	p.connector.Done()
}




// ------------------------------------------------

// commandBuilder формирует правильный срез аргументов конкретной команды
type commandBuilder struct {
	// процессор команд (отправка и получение результата/ошибки)
	proc processor
}

// NewCommander возвращает клиента, готового выполнять команды
func NewCommander(c connector) Commander {
	return &commandBuilder{
		proc: &commandProcessor{
			connector: c,
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
	go b.proc.process(cmd)

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
	go b.proc.process(cmd)

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
	go b.proc.process(cmd)

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
	go b.proc.process(cmd)

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