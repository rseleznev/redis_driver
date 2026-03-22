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

	commandsChan chan models.Command // канал для входящих команд приложения
}

// NewConn создает новое соединение и подключается к нему
func NewConn(opts models.Options) (*Conn, error) {
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

	conn := &Conn{
		Options: opts,
		socketFd: socketFd,
		epollFd: epollFd,
		commandsChan: make(chan models.Command),
	}

	// Запускаем Process в отдельной горутине
	go conn.Process()

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

// Process принимает команды от потоков, централизованно отправляет их
// и возвращает результат ждущему потоку
func (c *Conn) Process() {
	for cmd := range c.commandsChan {
		var data []byte

		for {
			// Отправляем команду
			err := message.Send(c.socketFd, cmd.SendingData)
			if err != nil {
				if err == models.ErrConnectionClosed {
					// Создаем и подключаем новый сокет
					newSocket, err := connection.Reconnect(c.Options, c.socketFd)
					if err != nil {
						panic(err)
					}
					c.socketFd = newSocket
				}
				panic(err)
			}

			// Читаем ответ
			data, err = message.Receive(c.socketFd)
			if err != nil {
				if errors.Is(err, models.ErrConnectionClosed) {
					// Создаем и подключаем новый сокет
					newSocket, err := connection.Reconnect(c.Options, c.socketFd)
					if err != nil {
						panic(err)
					}
					c.socketFd = newSocket

					// Повторяем цикл
					continue
				}	
			}
			break
		}
		// Возвращаем результат ждущей горутине
		cmd.ResultChan <- data
	}
}