package redis_driver

import (
	"errors"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/connection"
	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/message"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Conn представляет собой одно соединение с сервером
type Conn struct {
	models.Options
	
	socketFd int
	epollFd int
	proto uint8 // версия протокола RESP

	rcv message.Receiver
	commandsChan chan models.Command // канал для входящих команд приложения
}

// NewConn создает новое соединение и подключается к нему
func NewConn(opts models.Options) (*Conn, error) {
	initOptions(&opts)
	
	// Создаем epoll
	// В будущем не нужно будет создавать отдельный epoll для каждого соединения
	epollFd, err := epoll.New()
	if err != nil {
		return nil, err
	}

	// Создаем и подключаем сокет
	socketFd, err := connection.New(opts)
	if err != nil {
		return nil, err
	}

	// Receiver
	rcv := message.NewReceiver(socketFd, opts)

	conn := &Conn{
		Options: opts,
		socketFd: socketFd,
		epollFd: epollFd,
		rcv: rcv,
		commandsChan: make(chan models.Command),
	}

	// Запускаем process в отдельной горутине
	go conn.process()

	// Включаем протокол RESP3
	err = conn.Hello3()
	if err != nil {
		return nil, err
	}
	
	return conn, nil
}

// Close закрывает соединение
func (c *Conn) Close() {
	// Возможно лучше убрать сисколы поглубже
	syscall.Close(c.socketFd)
	syscall.Close(c.epollFd)

	close(c.commandsChan)
}

// process принимает команды от потоков, централизованно отправляет их
// и возвращает результат ждущему потоку
func (c *Conn) process() {
	for cmd := range c.commandsChan {
		var data []byte
		var err error

		for {
			// Отправляем команду
			err = message.Send(c.socketFd, cmd.SendingData, c.RetryAmount)
			if err != nil {
				if err == models.ErrConnectionClosed { // соединение закрыто, нужно переподключиться
					// Создаем и подключаем новый сокет
					newSocket, err := connection.Reconnect(c.Options, c.socketFd)
					if err != nil {
						cmd.ErrChan <- err
					}
					c.socketFd = newSocket

					continue
				}
				cmd.ErrChan <- err
				break
			}

			// Читаем ответ
			data, err = c.rcv.Receive()
			if err != nil {
				if errors.Is(err, models.ErrConnectionClosed) { // соединение закрыто, нужно переподключиться
					// Создаем и подключаем новый сокет
					newSocket, err := connection.Reconnect(c.Options, c.socketFd)
					if err != nil {
						cmd.ErrChan <- err
					}
					c.socketFd = newSocket

					continue
				}

				cmd.ErrChan <- err
			}
			break
		}
		// Возвращаем результат ждущей горутине
		cmd.ResultChan <- data
	}
}