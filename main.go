package main

import (
	"fmt"
	"syscall"
)

func main() {
	// Создаем сокет
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		fmt.Println("Ошибка создания сокета")
	} else {
		fmt.Println("Сокет создан!")
	}

	// Настраиваем таймауты
	timeoutSetting := syscall.Timeval{
		Sec: 5,
		Usec: 0,
	}
	syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &timeoutSetting)
	syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &timeoutSetting)

	// Включаем keep alive
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)

	// Настраиваем проверку keep alive
	syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, 300) // через 5 минут без активности...
	syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 60) // отправляется тестовый пакет, ждем 60 секунд...
	syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 5) // после 5 неудачных попыток соединение закрывается

	syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1) // отключаем задержки

	// Подключаемся
	var addr syscall.SockaddrInet4

	addr.Port = 6379
	addr.Addr = [4]byte{127, 0, 0, 1}

	// Подключение
	err = syscall.Connect(fd, &addr)
	if err != nil {
		fmt.Println("Ошибка подключения к серверу")
	} else {
		fmt.Println("Подключение успешно!")
	}

	defer syscall.Close(fd)

	testReq := []byte{'*', '1', '\r', '\n', '$', '4', '\r', '\n', 'P', 'I', 'N', 'G', '\r', '\n'}

	n, err := syscall.SendmsgN(fd, testReq, nil, nil, 0)
	if err != nil {
		fmt.Println("Ошибка отправки запроса")
	}
	fmt.Println(n)
}