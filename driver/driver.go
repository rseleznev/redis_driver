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

	workInProcess bool // есть ли ждущие потоки (временно)
	commandsChan chan models.Command // канал для входящих команд приложения
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

	// Запускаем Polling в отдельной горутине
	go conn.Polling()
	
	return conn, nil
}

// Close закрывает соединение
func (c *Conn) Close() {
	// Возможно лучше убрать сисколы поглубже
	syscall.Close(c.socketFd)
	syscall.Close(c.epollFd)

	// также закрывать Polling?
}

// Polling обрабатывает события и команды в бесконечном цикле
func (c *Conn) Polling() {
	// Обрабатываемая команда
	var currCmd *models.Command // временно
	
	// Бесконечный цикл
	// !также подумать, как правильно завершать при отключении!
	// !также подумать, в какой момент запускать. Плохо работать без команд и событий
	for {
		if c.workInProcess == true {
			// Проверяем новые события с таймаутом 0
			n, err := syscall.EpollWait(c.epollFd, epoll.WaitingEvents, 0)
			if err != nil {
				fmt.Println("ошибка ожидания epoll: ", err)
			}
			if n > 0 { // Пришли какие-то события
				// Проверки
				err = epoll.ProcessEvent(c.socketFd, epoll.WaitingEvents[0])
				if err != nil {
					fmt.Println(err)
				}

				// Читаем ответ
				data, err := receive.Message(c.socketFd)
				if err != nil {
					fmt.Println(err)
				}

				// Записываем в тот канал, где ждет инициирующий поток
				currCmd.ResultChan <- data // !разобраться, как распределять результаты между несколькими командами!
				currCmd = nil // готовы брать след команду
			}

			// Проверяем новые команды без блокировки
			select {
			case cmd := <-c.commandsChan:
				// Новая команда на исполнение

				// Отправляем команду
				err := send.Message(c.socketFd, cmd.SendingData)
				if err != nil {
					fmt.Println(err)
				}
				currCmd = &cmd // !тут нужно подумать, как быть с несколькими разными командами!

			default:
				continue
			}	
		}
	}
}

// Ping отправляет тестовую команду для проверки соединения
func (c *Conn) Ping() (string, error) {
	// Проверочная команда
	pingCommand := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'} // PING

	cmd := models.Command{
		Operation: "TEST",
		SendingData: pingCommand,
		ResultChan: make(chan []byte),
	}

	c.workInProcess = true
	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <- cmd.ResultChan
	c.workInProcess = false

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

	c.workInProcess = true
	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan
	c.workInProcess = false

	// Парсим сообщение
	parsed := message.Parse(data)
	fmt.Println(parsed)

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
	data := message.SerializeSetCommand("SET", "test", "fgdfjjjjhhuuhhhhhhhh", 112)
	for _, v := range data {
		fmt.Printf("Байт: %q \n", v)
	}

	// Отправляем команду на выполнение

	// Блокируемся и ждем результат
	
	return nil
}

func (c *Conn) GetValueByKey(key string) {
	getCommand := message.SerializeGetCommand(key)

	cmd := models.Command{
		Operation: "GET",
		SendingData: getCommand,
		ResultChan: make(chan []byte),
	}

	c.workInProcess = true
	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan
	c.workInProcess = false

	for i, v := range data {
		fmt.Printf("Байт: %q, индекс: %d \n", v, i)
	}
}