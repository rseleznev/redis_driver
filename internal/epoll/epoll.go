package epoll

import (
	"fmt"
	"errors"
	"syscall"
)

// События, которые хотим отслеживать
// !в пуле будет больше одного сокета
var WaitingEvents = make([]syscall.EpollEvent, 1)
// epoll всегда будет один независимо от кол-ва соединений
var epollFd int

var (
	ErrSocketEvent = errors.New("redis_driver: error event has happened on socket")
	ErrSocketHUPEvent = errors.New("redis_driver: HUP error event has happened on socket")
	ErrSocketRDHUPEvent = errors.New("redis_driver: RDHUP error event has happened on socket")
)

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

// Wait в цикле вызывает epoll с нулевым таймаутом, отдает управление,
// когда приходит одно из отслеживаемых событий в одном из отслеживаемых сокетов
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

// InitEventForSocket добавляет первичное событие в отслеживание для подключения сокета
func InitEventForSocket(socketFd int) error {
	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Добавляем событие в interest list
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, socketFd, &epollEvent)
	if err != nil {
		return fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено первичное событие в epoll")
	WaitingEvents[0] = epollEvent
	
	return nil
}

// AddIncomeEventForSocket добавляет в отслеживание входящее событие (что-то есть в буфере получения)
func AddIncomeEventForSocket(socketFd int) error {
	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Модифицируем событие в interest list
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_MOD, socketFd, &epollEvent)
	if err != nil {
		return fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено входящее событие в epoll")

	WaitingEvents[0] = epollEvent
	
	return nil
}

// AddOutcomeEventForSocket добавляет в отслеживание исходящее событие (буфер отправки пуст)
func AddOutcomeEventForSocket(socketFd int) error {
	// Создаем событие
	epollEvent := syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}

	// Модифицируем событие в interest list
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_MOD, socketFd, &epollEvent)
	if err != nil {
		return fmt.Errorf("ошибка добавления события в epoll: %w", err)
	}
	fmt.Println("Добавлено исходящее событие в epoll")

	WaitingEvents[0] = epollEvent
	
	return nil
}

// DeleteEventsForSocket удаляет все события сокета из списка отслеживания
func DeleteEventsForSocket(socketFd int) error {
	// Удаляем событие
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_DEL, socketFd, &WaitingEvents[0]) // events игнорируются
	if err != nil {
		return fmt.Errorf("ошибка удаления события в epoll: %w", err)
	}
	fmt.Println("Удалены события из epoll для сокета: ", socketFd)
	WaitingEvents[0] = syscall.EpollEvent{} // удаляем ненужное событие из переменной

	return nil
}

// ProcessEvent проверяет пришедшее событие и сокет
func ProcessEvent(socketFd int) error {
	// тут нужно получше сделать проверки!
	
	// Проверяем ошибку в сокете
	val, err := syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		fmt.Println("SO_ERROR, value: ", val)
		return err
	}

	var errs []error

	// Проверяем полученное событие
	if WaitingEvents[0].Events & syscall.EPOLLIN != 0 { // есть данные в буфере получения
		fmt.Println("Пришло событие EPOLLIN!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLOUT != 0 { // буфер отправки пуст
		fmt.Println("Пришло событие EPOLLOUT!")
	}
	if WaitingEvents[0].Events & syscall.EPOLLERR != 0 { // ошибка
		fmt.Println("Пришло событие EPOLLERR!")
		errs = append(errs, ErrSocketEvent)
	}
	if WaitingEvents[0].Events & syscall.EPOLLHUP != 0 { // соединение закрыто сервером
		fmt.Println("Пришло событие EPOLLHUP!")
		errs = append(errs, ErrSocketHUPEvent)
	}
	if WaitingEvents[0].Events & syscall.EPOLLRDHUP != 0 { // сервер закрыл запись
		fmt.Println("Пришло событие EPOLLRDHUP!")
		errs = append(errs, ErrSocketRDHUPEvent)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	
	return nil
}