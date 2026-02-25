package socket

import (
	"fmt"
	"syscall"
)

func ConnectNew(epollFd int) (int, error) {
	// Создаем сокет
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM | syscall.SOCK_NONBLOCK, syscall.IPPROTO_TCP)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания сокета: %w", err)
	}
	fmt.Println("Сокет создан!")

	// Настраиваем таймауты
	timeoutSetting := syscall.Timeval{
		Sec: 5,
		Usec: 0,
	}
	// !Изучить, какие тут могут быть ошибки
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

	// Подключение
	// Адрес другой стороны
	var addr syscall.SockaddrInet4
	addr.Port = 6379
	addr.Addr = [4]byte{127, 0, 0, 1}

	// Подключение сокета
	// !Убедиться, что можно не смотреть ошибку
	syscall.Connect(socketFd, &addr) // результат ловим через epoll

	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем запись в interest list
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, socketFd, &epollEvent)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено событие в epoll")

	// Ждем результат
	_, err = syscall.EpollWait(epollFd, []syscall.EpollEvent{epollEvent}, 100) // если за указанный таймаут события не будет, выполнение пойдет дальше
	if err != nil {
		return 0, fmt.Errorf("ошибка ожидания epoll: %w", err)
	}

	// Проверяем ошибку в сокете
	_, err = syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		fmt.Println("Ошибка в GetsockoptInt: ", err)
	}
	// Проверяем полученное событие
	if epollEvent.Events & syscall.EPOLLOUT != 0 {
		fmt.Println("Пришло событие EPOLLOUT!")
	}
	if epollEvent.Events & syscall.EPOLLERR != 0 {
		fmt.Println("Пришло событие EPOLLERR!")
	}

	fmt.Println("Сокет подключен!")

	return socketFd, nil
}