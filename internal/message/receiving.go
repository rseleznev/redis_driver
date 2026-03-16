package message

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/socket"
)

var (
	ErrMsgRcvTrunc = errors.New("redis_driver: received message truncated")
	ErrMsgRcvCTrunc = errors.New("redis_driver: received oob-message truncated")
	ErrConnClosed = errors.New("redis_driver: conn is closed by server")
)

func Receive(socketFd int) ([]byte, error) {
	var result []byte
	var err error

	for {
		result, err = tryReceive(socketFd)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				epoll.Wait()
				continue
			}
			if errors.Is(err, socket.ErrSocketClosed) {
				return nil, ErrConnClosed
			}
			// также надо проверять ErrMsgRcvTrunc и ErrMsgRcvCTrunc
		}
		break
	}
	
	// Проверки
	err = epoll.ProcessEvent(socketFd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Успешно получено")

	return result, err
}

// Прочитать ответ сервера
func tryReceive(socketFd int) ([]byte, error) {
	// Буфер для чтения
	buf := make([]byte, 1024) // 8192 как вариант

	n, _, coreFlags, _, err := syscall.Recvmsg(socketFd, buf, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	err = check(coreFlags)
	
	// Соединение закрыто сервером
	if n == 0 {
		return nil, socket.ErrSocketClosed
	}
	fmt.Println("Прочитано байт: ", n)

	return buf[:n], nil
}

// Проверка флагов ядра
func check(flags int) error {
	// В буфер влезло не все
	if flags & syscall.MSG_TRUNC != 0 {
		return ErrMsgRcvTrunc
	}
	// Доп проверка, не должна срабатывать
	if flags & syscall.MSG_CTRUNC != 0 {
		return ErrMsgRcvCTrunc
	}
	return nil
}