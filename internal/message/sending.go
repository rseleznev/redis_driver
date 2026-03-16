package message

import (
	"fmt"
	"errors"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
)

var (
	ErrMsgSndTrunc = errors.New("redis_driver: sended message truncated")
)

func Send(socket int, data []byte) error {
	var err error
	
	for {
		err = trySend(socket, data)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				epoll.Wait()
				continue
			}
			// также нужно проверять ErrMsgSndTrunc
			fmt.Println(err)
		}
		break
	}
	fmt.Println("Успешно отправлено")

	return err
}

func trySend(socket int, data []byte) error {
	// Системный вызов для отправки
	n, err := syscall.SendmsgN(socket, data, nil, nil, 0)
	if err != nil {
		// Возможные ошибки:
		// -данные не влезли в буфер
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	if n != len(data) {
		return ErrMsgSndTrunc
	}
	fmt.Println("Принято байт на отправку: ", n)
	
	return nil
}