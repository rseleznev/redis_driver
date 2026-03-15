package driver

import (
	"errors"
	"fmt"
	"strconv"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/message"
	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/internal/receive"
	"github.com/rseleznev/redis_driver/internal/send"
	"github.com/rseleznev/redis_driver/internal/socket"
)

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
// При успехе также запускается polling в бесконечном цикле отдельной горутины
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

	// Закидываем нужные события на отслеживание
	err = epoll.AddIncomeEvent(socketFd, epollFd)
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

	// также закрывать Process?
}

func (c *Conn) Process() {
	for cmd := range c.commandsChan {
		// Отправляем команду
		err := send.Message(c.socketFd, cmd.SendingData)
		if err != nil {
			// тут может быть ошибка, что буфер отправки полон и нужно ждать через epoll
			fmt.Println(err)
		}

		// Читаем ответ
		data, err := receive.Message(c.socketFd)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				// Ждем через epoll
				for {
					n, err := syscall.EpollWait(c.epollFd, epoll.WaitingEvents, 0)
					if err != nil {
						fmt.Println("ошибка ожидания epoll: ", err)
					}
					if n > 0 { // Пришли какие-то события
						break
					}
				}
				// Проверки
				err = epoll.ProcessEvent(c.socketFd, epoll.WaitingEvents[0])
				if err != nil {
					fmt.Println(err)
				}

				// Читаем ответ
				data, err = receive.Message(c.socketFd)
				if err != nil {
					// Если соединение закрыто сервером, нужно создать новое
					fmt.Println(err)
				}
			}
		}
		// Возвращаем результат ждущей горутине
		cmd.ResultChan <- data
	}
}

//-----------------------------------

// Ping отправляет тестовую команду для проверки соединения
func (c *Conn) Ping() (string, error) {
	// Проверочная команда
	pingCommand := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'} // PING

	cmd := models.Command{
		Operation: "TEST",
		SendingData: pingCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <- cmd.ResultChan

	// Парсим сообщение
	parsed := message.Parse(data)

	// Десериализация ответа
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.([]byte)
	if !ok {
		return "", errors.New("ошибка преобразования")
	}

	return string(result), nil
}

// Hello3 проверяет соединение и включает протокол RESP3
func (c *Conn) Hello3() (map[string]string, error) {
	helloCommand := []byte{
		'*', '2', '\r', '\n',
		'$', '5', '\r', '\n',
		'H', 'E', 'L', 'L', 'O', '\r', '\n', // HELLO
		'$', '1', '\r', '\n',
		'3', '\r', '\n',} // 3

	cmd := models.Command{
		Operation: "TEST",
		SendingData: helloCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	// Парсим сообщение
	parsed := message.Parse(data)

	// Десериализация ответа
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.(map[string]string)
	if !ok {
		return nil, errors.New("ошибка преобразования")
	}
	pv, _ := strconv.Atoi(result["proto"])
	c.proto = uint8(pv)
	
	return result, nil
}

// SetValueForKey устанавливает указанное значение value для указанного ключа key с длительностью хранения duration
func (c *Conn) SetValueForKey(key string, value any, duration int) error {
	setCommand := message.SerializeSetCommand(key, value, duration)

	cmd := models.Command{
		Operation: "SET",
		SendingData: setCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.([]byte)
	if !ok {
		return errors.New("ошибка преобразования")
	}
	fmt.Println(string(result))
	
	return nil
}

func (c *Conn) GetValueByKey(key string) any {
	getCommand := message.SerializeGetCommand(key)

	cmd := models.Command{
		Operation: "GET",
		SendingData: getCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)

	return deserialized
}