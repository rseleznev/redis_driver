package polling

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
)

type Epoller interface {
	Add(models.PollingUnit) error
	GetError() error
}

type epoll struct {
	// файловый дескриптор инстанса epoll
	fd int

	mu sync.Mutex
	err error // возможно лишнее

	// флаг, запущен ли поллинг
	polling bool

	// события, за которыми следим
	events []syscall.EpollEvent

	// сокеты, которые процессим и канал для возврата результата
	sockets map[int]models.PollingUnit
}

func NewPoller() (Epoller, error) {
	eFd, err := syscall.EpollCreate(1)
	if err != nil {
		// EINVAL size is not positive.
		// EINVAL (epoll_create1()) Invalid value specified in flags.

		// EMFILE The per-user limit on the number of epoll instances  imposed  by  /proc/sys/fs/epoll/max_user_instances
		// 		was encountered.  See epoll(7) for further details.
		// EMFILE The per-process limit on the number of open file descriptors has been reached.
		if errors.Is(err, syscall.EMFILE) {
			return nil, models.ErrTooManyFilesInProcess
		}

		// ENFILE The system-wide limit on the total number of open files has been reached.
		if errors.Is(err, syscall.ENFILE) {
			return nil, models.ErrTooManyFilesInSystem
		}

		// ENOMEM There was insufficient memory to create the kernel object.
		if errors.Is(err, syscall.ENOMEM) {
			return nil, models.ErrPollNoMemory
		}
		
		return nil, fmt.Errorf("polling creation err: %w", err)
	}

	return &epoll{
		fd: eFd,
		mu: sync.Mutex{},
	}, nil
}

// Add добавляет событие (юнит), которое нужно процессить
func (e *epoll) Add(unit models.PollingUnit) error {
	e.mu.Lock()

	defer e.mu.Unlock()

	// проверка, нет ли у нас переданного сокета в обработке
	if e.isSocketInProcess(unit.SocketFd) {
		return models.ErrSocketAlreadyAdded // вызывающий поток должен подождать, когда обработается текущее событие
	}

	// добавляем нужное событие в epoll_ctl и событие в events
	switch unit.EventType {

	// хотим узнать результат подключения к серверу
	case "connect":
		err := e.addConnectEvent(unit.SocketFd)
		if err != nil {
			return err
		}

	// хотим узнать факт получения входящего сообщения
	case "income":
		err := e.addIncomeEvent(unit.SocketFd)
		if err != nil {
			return err
		}

	// хотим узнать результат отправки своего сообщения
	case "outcome":
		err := e.addOutcomeEvent(unit.SocketFd)
		if err != nil {
			return err
		}

	default:
		return models.ErrPollUnknownEventType

	}

	e.addSocketInProcess(unit)

	// проверка, происходит ли поллинг. Если да - конец
	// Если нет - запускаем его
	if !e.isPolling() {
		e.startPolling()
		go e.wait()
	}

	return nil
}

// wait делает системный вызов epoll_wait с нулевым таймаутом
//
// Крутится, пока не получит события по всем ждущим сокетам
func (e *epoll) wait() {
	for {
		n, err := syscall.EpollWait(e.fd, e.events, 0)
		if err != nil {
			e.setError(err)
			e.pushError()
			break
		}
		if n > 0 { // Пришли какие-то события
			e.mu.Lock()

			if n == e.pollingLen() { // готовы все ожидаемые сокеты
				e.stopPolling()
				e.mu.Unlock()
				e.processEvents(n)
				
				break
			}
			e.mu.Unlock()
			go e.processEvents(n)
		}
	}
}

// processEvents обрабатывает полученные события, находит готовые сокеты и возвращает результаты ждущим потокам
func (e *epoll) processEvents(readySocketsLen int) {
	e.mu.Lock()

	defer e.mu.Unlock()

	readySockets := make(map[int]models.PollingResult, readySocketsLen)

	for i := 0; i < e.eventsLen(); i++ {
		var errs []error
		socketFd := int(e.events[i].Fd)

		// пытаемся получить ошибку по сокету
		_, err := syscall.GetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_ERROR)
		if err != nil {
			errs = append(errs, err)
		}

		// Проверяем полученное событие
		// проверяем на наличие событий с ошибками
		if e.events[i].Events & syscall.EPOLLERR != 0 { // ошибка
			errs = append(errs, models.ErrSocketEvent)
		}
		if e.events[i].Events & syscall.EPOLLHUP != 0 { // соединение закрыто сервером
			errs = append(errs, models.ErrSocketHUPEvent)
		}
		if e.events[i].Events & syscall.EPOLLRDHUP != 0 { // сервер закрыл запись
			errs = append(errs, models.ErrSocketRDHUPEvent)
		}

		// если по сокету есть ошибки, группируем их и закидываем в результирующий словарь
		if len(errs) > 0 {
			readySockets[socketFd] = models.PollingResult{
				Err: errors.Join(errs...),
			}

			continue
		}

		// проверяем корректные события
		if e.events[i].Events & syscall.EPOLLIN != 0 { // есть данные в буфере получения
			if e.getSocketEventType(socketFd) == "income" { // если ждем именно это событие
				readySockets[socketFd] = models.PollingResult{}

				continue
			}
			readySockets[socketFd] = models.PollingResult{
				Err: models.ErrPollDiffEventType,
			}
		}
		if e.events[i].Events & syscall.EPOLLOUT != 0 { // буфер отправки пуст
			if e.getSocketEventType(socketFd) == "outcome" { // если ждем именно это событие
				readySockets[socketFd] = models.PollingResult{}

				continue
			}
			readySockets[socketFd] = models.PollingResult{
				Err: models.ErrPollDiffEventType,
			}
		}
	}

	if len(readySockets) != readySocketsLen {
		e.setError(errors.New("not all expected sockets are ready")) // возможно не нужно проверять
	}

	for s, v := range readySockets {
		e.getSocketResultChan(s) <- v.Err // возвращаем результат. Ждущий поток может продолжить свое выполнение
	}
	e.deleteCompletedEpollEvents(readySockets) // удаляем завершенные события
}

func (e *epoll) setError(err error) {
	e.err = err
}

func (e *epoll) GetError() error {
	return e.err
}

// pushError информирует все ждущие потоки о глобальной ошибке epoll
func (e *epoll) pushError() {
	e.mu.Lock()

	defer e.mu.Unlock()

	for s := range e.sockets {
		e.getSocketResultChan(s) <- e.GetError()
	}
}


// ------------------------------------------------
// Методы, которые должны вызываться только под захваченным мьютексом

func (e *epoll) isSocketInProcess(socketFd int) bool {
	_, ok := e.sockets[socketFd]
	return ok
}

func (e *epoll) addSocketInProcess(unit models.PollingUnit) {
	e.sockets[unit.SocketFd] = unit
}

func (e *epoll) isPolling() bool {
	return e.polling
}

func (e *epoll) startPolling() {
	e.polling = true
}

func (e *epoll) stopPolling() {
	e.polling = false
}

func (e *epoll) pollingLen() int {
	return len(e.sockets)
}

func (e *epoll) eventsLen() int {
	return len(e.events)
}

func (e *epoll) newOutcomeEvent(socketFd int) syscall.EpollEvent {
	return syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}
}

func (e *epoll) newIncomeEvent(socketFd int) syscall.EpollEvent {
	return syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd: int32(socketFd),
		Pad: 0, // узнать, что это
	}
}

func (e *epoll) addConnectEvent(socketFd int) error {
	event := e.newOutcomeEvent(socketFd)
	
	err := syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, socketFd, &event)
	if err != nil {
		return e.handleEpollError(err)
	}
	e.addEpollEvent(event)
	
	return nil
}

func (e *epoll) addIncomeEvent(socketFd int) error {
	event := e.newIncomeEvent(socketFd)

	err := syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_MOD, socketFd, &event)
	if err != nil {
		return e.handleEpollError(err)
	}
	e.addEpollEvent(event)
	
	return nil
}

func (e *epoll) addOutcomeEvent(socketFd int) error {
	event := e.newOutcomeEvent(socketFd)

	err := syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_MOD, socketFd, &event)
	if err != nil {
		return e.handleEpollError(err)
	}
	e.addEpollEvent(event)
	
	return nil
}

func (e *epoll) getSocketResultChan(socketFd int) chan error {
	return e.sockets[socketFd].ResultChan
}

func (e *epoll) getSocketEventType(socketFd int) string {
	return e.sockets[socketFd].EventType
}

func (e *epoll) addEpollEvent(event syscall.EpollEvent) {
	e.events = append(e.events, event)
}

func (e *epoll) deleteCompletedEpollEvents(readySockets map[int]models.PollingResult) {
	newEvents := make([]syscall.EpollEvent, 0, e.eventsLen()-len(readySockets))
	
	for _, v := range e.events {
		if _, ok := readySockets[int(v.Fd)]; !ok {
			newEvents = append(newEvents, v)
		}
	}
	e.events = newEvents
}

func (e *epoll) handleEpollError(err error) error {
	// EBADF  epfd or fd is not a valid file descriptor.
	if errors.Is(err, syscall.EBADF) {
		return models.ErrPollBadFD
	}

	// EEXIST op was EPOLL_CTL_ADD, and the supplied file descriptor fd is already registered  with  this  epoll  in‐
	// 	stance.
	if errors.Is(err, syscall.EEXIST) {
		return models.ErrSocketAlreadyAdded
	}

	// EINVAL epfd  is  not an epoll file descriptor, or fd is the same as epfd, or the requested operation op is not
	// 	supported by this interface.
	// EINVAL An invalid event type was specified along with EPOLLEXCLUSIVE in events.
	// EINVAL op was EPOLL_CTL_MOD and events included EPOLLEXCLUSIVE.
	// EINVAL op was EPOLL_CTL_MOD and the EPOLLEXCLUSIVE flag has previously been applied to this epfd, fd pair.
	// EINVAL EPOLLEXCLUSIVE was specified in event and fd refers to an epoll instance.
	// ELOOP  fd refers to an epoll instance and this EPOLL_CTL_ADD operation would result  in  a  circular  loop  of
	// 	epoll instances monitoring one another or a nesting depth of epoll instances greater than 5.

	// ENOENT op was EPOLL_CTL_MOD or EPOLL_CTL_DEL, and fd is not registered with this epoll instance.
	if errors.Is(err, syscall.ENOENT) {
		return models.ErrSocketNotAdded
	}

	// ENOMEM There was insufficient memory to handle the requested op control operation.
	if errors.Is(err, syscall.ENOMEM) {
		return models.ErrNoMemory
	}

	// ENOSPC The  limit  imposed  by  /proc/sys/fs/epoll/max_user_watches  was  encountered while trying to register
	// 	(EPOLL_CTL_ADD) a new file descriptor on an epoll instance.  See epoll(7) for further details.
	// EPERM  The target file fd does not support epoll.  This error can occur if fd refers to, for example, a  regu‐
	// 	lar file or a directory.

	return fmt.Errorf("epoll_ctl err: %w", err)
}



// ------------------------------------------------