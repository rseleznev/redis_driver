package main

import (
	"fmt"
	"syscall"

	// "redis_driver/send"
	"redis_driver/epoll"
	"redis_driver/models"
	"redis_driver/socket"
	// "redis_driver/receive"
)

func main() {
	// Создаем сокет
	socketFd, err := socket.New()
	if err != nil {
		fmt.Println(err)
	}
	// Создаем epoll
	epollFd, err := epoll.New()
	if err != nil {
		fmt.Println(err)
	}

	// Адрес другой стороны
	var addr syscall.SockaddrInet4
	addr.Port = 6379
	addr.Addr = [4]byte{127, 0, 0, 1}
	
	// Подключение сокета
	err = syscall.Connect(*socketFd, &addr)
	if err != nil {
		fmt.Println("Блокировка на подключении сокета")
	}
	// создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd: int32(*socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем запись в interest list
	err = syscall.EpollCtl(*epollFd, syscall.EPOLL_CTL_ADD, *socketFd, &epollEvent)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Добавлена запись в epoll")

	defer syscall.Close(*socketFd)

	// Создаем очередь для ожидания событий
	eventQueue := []models.EpollEvent{}
	eventQueue = append(eventQueue, models.EpollEvent{
		Type: "connect",
		Event: epollEvent,
	})

	// Обрабатываем очередь событий
	for len(eventQueue) > 0 {
		event := eventQueue[0]

		switch event.Type {
		case "connect":
			_, err = syscall.EpollWait(*epollFd, []syscall.EpollEvent{event.Event}, 100) // здесь блокировка на каждом событии
			if err != nil {
				fmt.Println("ошибка ожидания epoll: ", err)
			}
			fmt.Println("epoll настроен!?")
			eventQueue = eventQueue[:len(eventQueue)-1]
		}
	}

	// // Проверочная команда
	// testReq := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'}

	// // Отправка сообщения
	// // Здесь блокировка при отправке
	// err = send.Message(*socketFd, testReq)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // Читаем ответ
	// // Здесь блокировка при чтении
	// response, err := receive.Message(*socketFd)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(response)
}