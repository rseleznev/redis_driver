package main

import (
	"fmt"
	"log"
	"syscall"
)

func main() {
	// Создаем сокет
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		fmt.Println("Ошибка создания сокета", err)
	} else {
		fmt.Println("Сокет создан!")
	}

	// Создаем epoll
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		fmt.Println("Ошибка создания epoll", err)
	} else {
		fmt.Println("epoll создан")
	}

	// Создаем epoll_event
	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем запись в interest list
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, socketFd, &event)
	if err != nil {
		fmt.Println("Ошибка добавления записи в epoll", err)
	} else {
		fmt.Println("Добавлена запись в epoll")
	}

	n, err := syscall.EpollWait(epollFd, []syscall.EpollEvent{event}, 100)
	if err != nil {
		fmt.Println("Ошибка ожидания epoll", err)
	} else {
		fmt.Println("epoll настроен!?")
		fmt.Println(n)
	}

	// Настраиваем таймауты
	timeoutSetting := syscall.Timeval{
		Sec: 5,
		Usec: 0,
	}
	syscall.SetsockoptTimeval(socketFd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &timeoutSetting)
	syscall.SetsockoptTimeval(socketFd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &timeoutSetting)

	// Включаем keep alive
	syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)

	// Настраиваем проверку keep alive
	// специфично для Linux!
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, 300) // через 5 минут без активности...
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 60) // отправляется тестовый пакет, ждем 60 секунд...
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 5) // после 5 неудачных попыток соединение закрывается

	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1) // отключаем задержки

	// Подключаемся
	var addr syscall.SockaddrInet4

	addr.Port = 6379
	addr.Addr = [4]byte{127, 0, 0, 1}

	// Подключение
	err = syscall.Connect(socketFd, &addr)
	if err != nil {
		fmt.Println("Ошибка подключения к серверу", err)
	} else {
		fmt.Println("Подключение успешно!")
	}

	defer syscall.Close(socketFd)
	defer syscall.Close(epollFd)

	// Проверочная команда
	testReq := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'}

	n, err = syscall.SendmsgN(socketFd, testReq, nil, nil, 0)
	if err != nil {
		fmt.Println("Ошибка отправки запроса", err)
	}
	if n != len(testReq) {
		fmt.Println("Не все данные отправлены!")
	} else {
		fmt.Println("Принято байт на отправку: ", n)
	}

	// Читаем ответ
	buf := make([]byte, 1024) // 8192

	n, _, coreFlags, _, err := syscall.Recvmsg(socketFd, buf, nil, 0)
	if err != nil {
		fmt.Println("Ошибка чтения ответа", err)
		log.Fatal()
	}
	// Проверка флагов ядра
	// В буфер влезло не все
	if coreFlags & syscall.MSG_TRUNC != 0 {
		fmt.Println("Данные не влезли в буфер!")
	}
	// Доп проверка, не должна срабатывать
	if coreFlags & syscall.MSG_CTRUNC != 0 {
		fmt.Println("Обрезаны oob-данные")
}

	fmt.Println("Прочитано байт: ", n)
	ParseResponse(buf[:n])
}

func ParseResponse(input []byte) string {
	var result string

	for _, v := range input {
		fmt.Printf("Байт: %q \n", v)
	}

	return result
}