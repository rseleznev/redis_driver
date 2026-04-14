package connection

import (
	"errors"
	"sync"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
)

type socketer interface {
	GetSocketFd() int
	Connect(*models.Options) error
}

type poller interface {
	Add(models.PollingUnit) error
	GetError() error
}

type sender interface {
	Send() error
}

type receiver interface {
	Receive() error
}

type Connection struct {
	socketFd int
	opts *models.Options
	mu sync.Mutex

	// фабрика для создания нужных объектов, 
	// сохраняем в структуру для использования в будущем
	factory Factory

	// интерфейсы
	poller poller
	socket socketer
	sender sender
	receiver receiver

	// буферы
	sendBuf *models.SendBuf
	recvBuf *models.RecvBuf

	// флаг, есть ли сейчас команда в обработке
	// пока одна команда не обработается до конца, новые не принимаются
	processing bool
}

func NewConnection(opts *models.Options) (*Connection, error) {
	var f Factory
	
	// создаем сокет
	s, err := f.NewSocket(opts)
	if err != nil {
		return nil, err
	}

	// создаем поллер
	p, err := f.NewPoller()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		socketFd: s.GetSocketFd(),
		opts: opts,
		mu: sync.Mutex{},
		factory: f,
		poller: p,
		socket: s,
	}

	// подключаем сокет
	err = conn.connect()
	if err != nil {
		return nil, err
	}

	// создаем буферы
	conn.sendBuf = &models.SendBuf{
		SocketFd: s.GetSocketFd(),
		Buf: make([]byte, 0, 50), // брать из опций
		PrevCap: 50,
	}
	conn.recvBuf = &models.RecvBuf{
		SocketFd: s.GetSocketFd(),
		Buf: make([]byte, 0, opts.ReceiveBufAvgLen),
		PrevCap: opts.ReceiveBufAvgLen,
	}
	
	return conn, nil
}

// connect выполняет подключение к серверу
func (c *Connection) connect() error {
	err := c.socket.Connect(c.opts)
	if err != nil {
		// сокет в неблокирующем режиме и нужно поллить
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EINPROGRESS) {
			err = c.poll("connect")
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	
	return nil
}

// poll поллит нужное событие, возвращает результат поллинга:
//
// - nil при событии connect означает, что соединение установлено
//
// - nil при событии income означает, что есть входящие данные
//
// - nil при событии outcome означает, что исходящие данные отправлены
func (c *Connection) poll(eventType string) error {
	pUnit := models.PollingUnit{
		SocketFd: c.socketFd,
		EventType: eventType,
		ResultChan: make(chan error),
	}

	var err error

	for {
		err = c.poller.Add(pUnit)
		if err != nil {
			// сокет уже поллится, крутимся в цикле и пробуем отдать наше событие
			if err == models.ErrSocketAlreadyAdded {
				continue
			}

			// какие еще ошибки тут могут быть:
			// - асинхронная ошибка
			// - неизвестный тип события
			return err
		}
		break
	}

	// блокируемся на чтении результата поллинга
	err = <-pUnit.ResultChan
	if err != nil {
		return err
	}

	return nil
}

// GetSendBuf отдает буфер отправки для заполнения данными если не процессится другая команда
func (c *Connection) GetSendBuf() (*models.SendBuf, error) {
	c.mu.Lock()
	
	defer c.mu.Unlock()

	if c.isProcessing() {
		return nil, models.ErrConnectionCmdInProcess
	}
	c.startProcessing()

	// очистка буфера отправки
	c.drainSendBuf()
	
	return c.sendBuf, nil
}

// CancelProcessing отменяет начатую операцию, очищает буфер отправки для следующей операции
func (c *Connection) CancelProcessing() {
	c.mu.Lock()

	defer c.mu.Unlock()

	// очистка буфера получения
	c.DrainRecvBuf(c.recvBuf)

	c.stopProcessing()
}

// SendAndReceive отправляет данные, ждет результат и возвращает его
func (c *Connection) SendAndReceive(*models.SendBuf) (*models.RecvBuf, error) {
	// отправка данных

	// поллинг

	// очистка буфера отправки

	// получение

	// поллинг
	
	return c.recvBuf, nil
}

// DrainRecvBuf очищает буфер получения, скидывает флаг процессинга.
// Вызывается, когда команда получила свои данные
func (c *Connection) DrainRecvBuf(_ *models.RecvBuf) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.stopProcessing()
	c.drainRecvBuf()
}

func (c *Connection) drainRecvBuf() {
	c.recvBuf.Buf = c.recvBuf.Buf[:0]
}

func (c *Connection) drainSendBuf() {
	c.sendBuf.Buf = c.sendBuf.Buf[:0]
}


// ------------------------------------------------
// Методы, которые должны вызываться только под захваченным мьютексом

func (c *Connection) isProcessing() bool {
	return c.processing
}

func (c *Connection) startProcessing() {
	c.processing = true
}

func (c *Connection) stopProcessing() {
	c.processing = false
}