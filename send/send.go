package send

import (
	"fmt"
	"errors"
	"syscall"
)

// Отправить сообщение на сервер
func Message(socket int, data []byte) error {
	// Системный вызов для отправки
	n, err := syscall.SendmsgN(socket, data, nil, nil, 0)
	if err != nil {
		// Возможные ошибки:
		// -данные не влезли в буфер
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	if n != len(data) {
		return errors.New("не все данные отправлены")
	}
	fmt.Println("Принято байт на отправку: ", n)
	
	return nil
}