package driver

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/message"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/socket"
)

// Conn представляет собой одно соединение с сервером
type Conn struct {
	socketFd int
	epollFd int
	redisIp [4]byte // нужно будет брать из конфига
	redisPort int // нужно будет брать из конфига
	proto uint8 // версия протокола RESP

	// настройки таймаутов
	// размер буферов

	commandsChan chan models.Command // канал для входящих команд приложения (возможно лучше сделать указателем)
}

// NewConn создает новое соединение и подключается к нему
func NewConn(ip [4]byte, port int) (*Conn, error) {
	// Создаем epoll
	// В будущем не нужно будет создавать отдельный epoll для каждого соединения
	epollFd, err := epoll.New()
	if err != nil {
		return nil, err
	}

	// Создаем и подключаем сокет
	socketFd, err := socket.ConnectNew(ip, port, epollFd)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		socketFd: socketFd,
		epollFd: epollFd,
		redisIp: ip,
		redisPort: port,
		commandsChan: make(chan models.Command),
	}

	// Запускаем Process в отдельной горутине
	go conn.Process()
	
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
		fmt.Println("Socket fd: ", c.socketFd)
		// Отправляем команду
		err := message.Send(c.socketFd, cmd.SendingData)
		if err != nil {
			fmt.Println(err)
		}

		// Читаем ответ
		data, err := message.Receive(c.socketFd)
		if err != nil {
			if errors.Is(err, message.ErrConnClosed) {
				fmt.Println("conn closed")
				newSocket, err := socket.ConnectNew(c.redisIp, c.redisPort, c.epollFd)
				if err != nil {
					fmt.Println(err)
				}
				c.socketFd = newSocket
				fmt.Println("Socket fd: ", c.socketFd)
			}
		}
		// Возвращаем результат ждущей горутине
		cmd.ResultChan <- data
	}
}