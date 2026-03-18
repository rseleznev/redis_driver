package driver

import (
	"fmt"
	"errors"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/message"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Ping отправляет тестовую команду для проверки соединения
func (c *Conn) Ping() (string, error) {
	fmt.Println("Вызвана команда Ping")
	pingCommand := message.SerializeCommand("PING")

	cmd := models.Command{
		Operation: "TEST",
		SendingData: pingCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Process не заберет команду

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
	fmt.Println("Команда Ping завершена")

	return string(result), nil
}

// Hello3 проверяет соединение и включает протокол RESP3
func (c *Conn) Hello3() error {
	fmt.Println("Вызвана команда Hello3")
	helloCommand := message.SerializeCommand("HELLO", "3")

	cmd := models.Command{
		Operation: "TEST",
		SendingData: helloCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Process не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	// Парсим сообщение
	parsed := message.Parse(data)

	// Десериализация ответа
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.(map[string]string)
	if !ok {
		return errors.New("ошибка преобразования")
	}
	pv, _ := strconv.Atoi(result["proto"])
	c.proto = uint8(pv)
	fmt.Println(result)
	fmt.Println("Команда Hello3 завершена")
	
	return nil
}

// SetValueForKey устанавливает указанное значение value для указанного ключа key с длительностью хранения duration.
//
// Параметр value должен быть строкой или срезом байт.
// Передача duration = 0 означает, что значение будет храниться без ограничения по времени
func (c *Conn) SetValueForKey(key string, value any, duration int) error {
	fmt.Println("Вызвана команда SetValueForKey")
	var setCommand []byte
	if duration == 0 {
		setCommand = message.SerializeCommand("SET", key, value)
	} else {
		durString := strconv.Itoa(duration)
		setCommand = message.SerializeCommand("SET", key, value, "EX", durString)
	}

	cmd := models.Command{
		Operation: "SET",
		SendingData: setCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Process не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.([]byte)
	if !ok {
		return errors.New("ошибка преобразования")
	}
	fmt.Println(string(result))
	fmt.Println("Команда SetValueForKey завершена")
	
	return nil
}

// GetValueByKey возвращает значение по ключу key
//
// Возвращается строка или срез байт
func (c *Conn) GetValueByKey(key string) any {
	fmt.Println("Вызвана команда GetValueByKey")
	getCommand := message.SerializeCommand("GET", key)

	cmd := models.Command{
		Operation: "GET",
		SendingData: getCommand,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Process не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan
	for _, v := range data {
		fmt.Printf("Байт: %q \n", v)
	}

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)
	fmt.Println("Команда GetValueByKey завершена")

	return deserialized
}

func (c *Conn) IncorrectTestCommand() {
	fmt.Println("Вызвана команда IncorrectTestCommand")
	command := []byte{
		'*', '2', '\r', '\n',
		'$', '3', '\r', '\n',
		'S', 'E', 'T', '\r', '\n',
		'%', '1', '\r', '\n',
		'$', '4', '\r', '\n',
		't', 'e', 's', 't', '\r', '\n',
		'$', '2', '\r', '\n',
		'v', 'h', '\r', '\n',
	}

	cmd := models.Command{
		Operation: "SET",
		SendingData: command,
		ResultChan: make(chan []byte),
	}

	c.commandsChan <- cmd // блокировка, пока Process не заберет команду

	// Блокируемся и ждем результат
	data := <-cmd.ResultChan

	// Парсим сообщение
	parsed := message.Parse(data)

	// Десериализация ответа
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.([]byte)
	if !ok {
		fmt.Println("ошибка преобразования")
		return
	}
	fmt.Println(string(result))
	fmt.Println("Команда IncorrectTestCommand завершена")
}