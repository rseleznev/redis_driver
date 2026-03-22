package message

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Receive получает сообщение по указанному socket
func Receive(socket int) ([]byte, error) {
	var result []byte
	var err error

	for {
		result, err = tryReceive(socket)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				epoll.Wait()
				continue
			}
			if errors.Is(err, models.ErrConnectionClosed) {
				return nil, models.ErrConnectionClosed
			}
			// также надо проверять ErrMsgRcvTrunc и ErrMsgRcvCTrunc
		}
		break
	}
	
	// Проверки
	err = epoll.ProcessEvent(socket)
	if err != nil {
		fmt.Println(err)
	}

	return result, err
}

// tryReceive делает одну попытку прочитать ответ сервера
// и выполняет проверки, если ответ есть
func tryReceive(socket int) ([]byte, error) {
	// Буфер для чтения
	buf := make([]byte, 1024) // 8192 как вариант

	// Читаем ответ
	n, _, coreFlags, _, err := syscall.Recvmsg(socket, buf, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем флаги ядра
	if coreFlags & syscall.MSG_TRUNC != 0 {
		return nil, models.ErrRecvMsgTrunc
	}
	// Доп проверка, не должна срабатывать
	if coreFlags & syscall.MSG_CTRUNC != 0 {
		return nil, models.ErrRecvMsgCTrunc
	}
	
	// Соединение закрыто сервером
	if n == 0 {
		return nil, models.ErrConnectionClosed
	}
	fmt.Println("Прочитано байт: ", n)

	return buf[:n], nil
}