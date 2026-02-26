package main

import (
	"fmt"
	"syscall"

	"redis_driver/epoll"
	"redis_driver/receive"
	"redis_driver/send"
	"redis_driver/socket"
	"redis_driver/serialization"
)

var EpollFd int

func main() {
	// Создаем epoll
	epollFd, err := epoll.New()
	if err != nil {
		fmt.Println(err)
	}
	EpollFd = epollFd

	// Создаем сокет
	socketFd, err := socket.ConnectNew(EpollFd)
	if err != nil {
		fmt.Println(err)
	}

	defer syscall.Close(socketFd) // в конце нужно будет убрать
	defer syscall.Close(EpollFd) // в конце нужно будет убрать?

	test, err := Ping(socketFd)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(test)
}

// Проверочная команда
// Работает только в одном потоке
func Ping(socketFd int) (string, error) {
	// Проверочная команда
	pingCommand := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'} // PING
	
	// Отправляем команду
	err := send.Message(socketFd, pingCommand)
	if err != nil {
		fmt.Println(err)
	}

	// Ждем результат, внутри блокировка!
	err = epoll.Wait(socketFd, EpollFd)
	if err != nil {
		fmt.Println(err)
	}

	// Читаем ответ
	data, err := receive.Message(socketFd)
	if err != nil {
		fmt.Println(err)
	}

	// Десериализация ответа
	result := serialization.Decode(data)

	return result, nil
}

// Заготовка на будущее
func GetValueByKey(key string) string {
	return ""
}

// Заготовка на будущее
func SetValueForKey(key, value string) string {
	return ""
}