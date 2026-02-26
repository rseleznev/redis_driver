package epoll

import (
	"fmt"
	"syscall"
)

// Создаем epoll
func New() (int, error) {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания epoll: %w", err)
	}
	fmt.Println("epoll создан")
	return epollFd, nil
}

func Wait(socketFd, epollFd int) error {
	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем событие в interest list
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_MOD, socketFd, &epollEvent)
	if err != nil {
		return fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено событие в epoll")

	// Ждем результат
	_, err = syscall.EpollWait(epollFd, []syscall.EpollEvent{epollEvent}, 100) // если за указанный таймаут события не будет, выполнение пойдет дальше
	if err != nil {
		return fmt.Errorf("ошибка ожидания epoll: %w", err)
	}

	// Проверяем ошибку в сокете
	_, err = syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		return fmt.Errorf("ошибка в GetsockoptInt: %w", err)
	}
	// Проверяем полученное событие
	if epollEvent.Events & syscall.EPOLLIN != 0 {
		fmt.Println("Пришло событие EPOLLIN!")
	}
	if epollEvent.Events & syscall.EPOLLERR != 0 {
		fmt.Println("Пришло событие EPOLLERR!")
	}

	return nil
}