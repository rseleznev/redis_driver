package socket

import (
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
)

// ConnectNew создает и подключает новый сокет
func ConnectNew(ip [4]byte, port, epollFd int) (int, error) {
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

	// Настройки буферов

	// Подключение
	// Адрес сервера
	var addr syscall.SockaddrInet4
	addr.Port = port // нужно будет сделать валидацию
	addr.Addr = ip // нужно будет сделать валидацию

	// Подключение сокета
	// !Убедиться, что можно не смотреть ошибку
	syscall.Connect(socketFd, &addr) // результат ловим через epoll

	// Добавляем в epoll входящие и исходящие события
	err = epoll.AddEvents(socketFd, epollFd)

	// Ждем результат
	epoll.Wait()

	// Проверка события
	err = epoll.ProcessEvent(socketFd)
	if err != nil {
		return 0, err
	}
	fmt.Println("Сокет подключен!")

	return socketFd, nil
}