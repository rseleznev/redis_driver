package redis_driver

import (
	"errors"
	"strconv"

	"github.com/rseleznev/redis_driver/internal/message"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Ping отправляет тестовую команду для проверки соединения
func (c *Conn) Ping() (string, error) {
	pingCommand, err := message.SerializeCommand("PING")
	if err != nil {
		return "", err
	}

	cmd := models.Command{
		Operation:   "TEST",
		SendingData: pingCommand,
		ResultChan:  make(chan []byte),
		ErrChan:     make(chan error),
	}

	c.commandsChan <- cmd // блокировка, пока process не заберет команду

	// Блокируемся и ждем результат
	var data []byte

	select {
	case data = <-cmd.ResultChan:

	case err = <-cmd.ErrChan:
		return "", err

	}

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
func (c *Conn) Hello3() error {
	helloCommand, err := message.SerializeCommand("HELLO", "3")
	if err != nil {
		return err
	}

	cmd := models.Command{
		Operation:   "TEST",
		SendingData: helloCommand,
		ResultChan:  make(chan []byte),
		ErrChan:     make(chan error),
	}

	c.commandsChan <- cmd // блокировка, пока process не заберет команду

	// Блокируемся и ждем результат
	var data []byte

	select {
	case data = <-cmd.ResultChan:

	case err = <-cmd.ErrChan:
		return err

	}

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

	return nil
}

// SetValueForKey устанавливает указанное значение value для указанного ключа key с длительностью хранения duration.
//
// Параметр value должен быть строкой или срезом байт.
// Передача duration = 0 означает, что значение будет храниться без ограничения по времени
func (c *Conn) SetValueForKey(key string, value any, duration int) error {
	var setCommand []byte
	var err error
	if duration == 0 {
		setCommand, err = message.SerializeCommand("SET", key, value)
	} else {
		durString := strconv.Itoa(duration)
		setCommand, err = message.SerializeCommand("SET", key, value, "EX", durString)
	}
	if err != nil {
		return err
	}

	cmd := models.Command{
		Operation:   "SET",
		SendingData: setCommand,
		ResultChan:  make(chan []byte),
		ErrChan:     make(chan error),
	}

	c.commandsChan <- cmd // блокировка, пока process не заберет команду

	// Блокируемся и ждем результат
	var data []byte

	select {
	case data = <-cmd.ResultChan:

	case err = <-cmd.ErrChan:
		return err

	}

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)

	if err, ok := deserialized.(error); ok {
		return err
	}

	return nil
}

// GetValueByKey возвращает значение по ключу key
//
// Возвращается строка или срез байт
func (c *Conn) GetValueByKey(key string) (any, error) {
	getCommand, err := message.SerializeCommand("GET", key)
	if err != nil {
		return nil, err
	}

	cmd := models.Command{
		Operation:   "GET",
		SendingData: getCommand,
		ResultChan:  make(chan []byte),
		ErrChan:     make(chan error),
	}

	c.commandsChan <- cmd // блокировка, пока process не заберет команду

	// Блокируемся и ждем результат или ошибку
	var data []byte

	select {
	case data = <-cmd.ResultChan:

	case err = <-cmd.ErrChan:
		return nil, err

	}

	parsed := message.Parse(data)
	deserialized := message.Deserialize(parsed)

	if err, ok := deserialized.(error); ok {
		return nil, err
	}

	return deserialized, nil
}

// Проверочная команда
func (c *Conn) incorrectTestCommand(input []byte) error {

	cmd := models.Command{
		Operation:   "SET",
		SendingData: input,
		ResultChan:  make(chan []byte),
		ErrChan:     make(chan error),
	}

	c.commandsChan <- cmd // блокировка, пока process не заберет команду

	// Блокируемся и ждем результат
	var data []byte

	select {
	case data = <-cmd.ResultChan:

	case err := <-cmd.ErrChan:
		return err

	}

	// Парсим сообщение
	parsed := message.Parse(data)

	// Десериализация ответа
	deserialized := message.Deserialize(parsed)

	result, ok := deserialized.(error)
	if ok {
		return result
	}

	return nil
}
