package epoll

import (
	"fmt"
	"syscall"
)

// Создаем epoll
func New() (*int, error) {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания epoll: %w", err)
	}
	fmt.Println("epoll создан")
	return &epollFd, nil
}

func Wait(sFD, eFD int) error {
	// Создаем epoll_event
	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		Fd: int32(sFD),
		Pad: 0, // узнать, что это
	}

	// Добавляем запись в interest list
	err := syscall.EpollCtl(eFD, syscall.EPOLL_CTL_ADD, sFD, &event)
	if err != nil {
		return fmt.Errorf("ошибка добавления записи в epoll: %w", err)
	}
	fmt.Println("Добавлена запись в epoll")

	_, err = syscall.EpollWait(eFD, []syscall.EpollEvent{event}, 100)
	if err != nil {
		return fmt.Errorf("ошибка ожидания epoll: %w", err)
	}
	fmt.Println("epoll настроен!?")

	return nil
}