package epoll

import (
	"fmt"
	"syscall"
)

// События, которые хотим отслеживать
var WaitingEvents = make([]syscall.EpollEvent, 1)
var epollFd int

// New создает новый инстанс epoll
func New() (int, error) {
	eFd, err := syscall.EpollCreate(1)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания epoll: %w", err)
	}
	fmt.Println("epoll создан")
	epollFd = eFd
	return eFd, nil
}

func Wait() {
	for {
		n, err := syscall.EpollWait(epollFd, WaitingEvents, 0)
		if err != nil {
			fmt.Println("ошибка ожидания epoll: ", err)
		}
		if n > 0 { // Пришли какие-то события
			break
		}
	}
}

func AddEvents(socketFd, epollFd int) error {
	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем событие в interest list
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, socketFd, &epollEvent)
	if err != nil {
		return fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено событие в epoll")

	WaitingEvents[0] = epollEvent
	
	return nil
}

// ProcessEvent проверяет пришедшее событие и сокет
// возможно лучше вынести в отдельный пакет проверок
func ProcessEvent(socketFd int) error {
	// тут нужно получше сделать проверки!
	
	// Проверяем ошибку в сокете
	_, err := syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		return fmt.Errorf("ошибка в GetsockoptInt: %w", err)
	}
	// Проверяем полученное событие
	if WaitingEvents[0].Events & syscall.EPOLLIN != 0 { // входящее сообщение
		fmt.Println("Пришло событие EPOLLIN!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLOUT != 0 {
		fmt.Println("Пришло событие EPOLLOUT!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLERR != 0 { // ошибка
		fmt.Println("Пришло событие EPOLLERR!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLHUP != 0 { // соединение закрыто сервером
		fmt.Println("Пришло событие EPOLLHUP!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLRDHUP != 0 { // сервер закрыл запись
		fmt.Println("Пришло событие EPOLLRDHUP!")
	}
	
	return nil
}