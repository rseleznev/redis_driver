package epoll

import (
	"fmt"
	"syscall"
)

// New создает новый инстанс epoll
func New() (int, error) {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания epoll: %w", err)
	}
	fmt.Println("epoll создан")
	return epollFd, nil
}

// Срез событий, которые хотим отслеживать. По факту только EPOLLIN
// завязано на конкретный сокет
var WaitingEvents = make([]syscall.EpollEvent, 1)

// AddIncomeEvent добавляет событие EPOLLIN в interest list указанного epollFd для socketFd
func AddIncomeEvent(socketFd, epollFd int) error {
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

	WaitingEvents[0] = epollEvent
	
	return nil
}

// ProcessEvent проверяет пришедшее событие и сокет
// возможно лучше вынести в отдельный пакет проверок
func ProcessEvent(socketFd int, event syscall.EpollEvent) error {
	// тут нужно получше сделать проверки!
	
	// Проверяем ошибку в сокете
	_, err := syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		return fmt.Errorf("ошибка в GetsockoptInt: %w", err)
	}
	// Проверяем полученное событие
	if event.Events & syscall.EPOLLIN != 0 {
		fmt.Println("Пришло событие EPOLLIN!")
	}
	if event.Events & syscall.EPOLLERR != 0 {
		fmt.Println("Пришло событие EPOLLERR!")
	}
	
	return nil
}