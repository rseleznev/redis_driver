package main

import (
	"fmt"
	"syscall"

	"redis_driver/send"
	"redis_driver/socket"
	"redis_driver/receive"
)

func main() {
	// Создаем epoll
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		fmt.Println("ошибка создания epoll: ", err)
	}
	// Создаем сокет
	socketFd, err := socket.ConnectNew(epollFd)
	if err != nil {
		fmt.Println(err)
	}

	defer syscall.Close(socketFd)

	// Проверочная команда
	testReq := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'}

	// Отправка сообщения
	// Здесь блокировка при отправке
	err = send.Message(socketFd, testReq)
	if err != nil {
		fmt.Println(err)
	}

	// Читаем ответ
	// Здесь блокировка при чтении
	response, err := receive.Message(socketFd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}