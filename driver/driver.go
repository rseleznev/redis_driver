package driver

import (
	"fmt"
	"syscall"

	"redis_driver/send"
	"redis_driver/epoll"
	"redis_driver/models"
	"redis_driver/socket"
	"redis_driver/receive"
	"redis_driver/serialization"
)

type Conn struct {
	socketFd int
	epollFd int
	redisIp [4]byte // нужно будет брать из конфига
	redisPort int // нужно будет брать из конфига

	// настройки таймаутов
	// размер буферов

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
	// Очередь команд
	var cmdsQueue = make([]models.Command, 0 ,5) // временно
	
	// Бесконечный цикл
	// !также подумать, как правильно завершать при отключении!
	for {
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
			for _, v := range cmdsQueue {
				v.ResultChan <- data // !разобраться, как распределять результаты между несколькими командами!
			}
			cmdsQueue = cmdsQueue[:0] // очищаем очередь команд
		}

		// Проверяем новые команды без блокировки
		select {
		case cmd := <- c.commandsChan:
			// Новая команда на исполнение

			if cmd.Operation == "TEST" {
				// Отправляем команду
				err := send.Message(c.socketFd, cmd.Key)
				if err != nil {
					fmt.Println(err)
				}
				cmdsQueue = append(cmdsQueue, cmd) // !тут нужно подумать, как быть с несколькими разными командами!
			}
		default:
			continue
		}
	}
}

// Ping отправляет тестовую команду для проверки соединения
func (c *Conn) Ping() (string, error) {
	// Проверочная команда
	pingCommand := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'} // PING

	cmd := models.Command{
		Operation: "TEST",
		Key: pingCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Polling не заберет команду

	// Блокируемся и ждем результат
	data := <- cmd.ResultChan

	// Десериализация ответа
	result := serialization.Decode(data)

	return result, nil
}

// SetValueForKey устанавливает указанное значение value для указанного ключа key с длительностью хранения durr
func (c *Conn) SetValueForKey(key, value string, durr int) error {
	// Формируем команду в формате RESP

	// Отправляем команду на выполнение

	// Блокируемся и ждем результат
	
	return nil
}