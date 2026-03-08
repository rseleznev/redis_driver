package receive

import (
	"fmt"
	"syscall"
)

// Прочитать ответ сервера
func Message(socketFd int) ([]byte, error) {
	// Буфер для чтения
	buf := make([]byte, 1024) // 8192 как вариант

	n, _, coreFlags, _, err := syscall.Recvmsg(socketFd, buf, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
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

	return buf[:n], nil
}