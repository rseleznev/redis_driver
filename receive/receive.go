package receive

import (
	"fmt"
	"syscall"
	"redis_driver/parser"
)

// Читаем ответ
func Message(socket int) (string, error) {
	buf := make([]byte, 1024) // 8192 как вариант

	n, _, coreFlags, _, err := syscall.Recvmsg(socket, buf, nil, 0)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
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
	result := parser.ParseResponse(buf[:n])

	return result, nil
}