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

	// фабрика для создания нужных объектов, сохраняем в структуру для возможности создания в будущем
	f Factory

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
		f: f,
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
		if errors.Is(err, syscall.EAGAIN) {
			err = c.poll("connect")
			if err != nil {
				return err
			}
			
			return nil
		}
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

	err := c.poller.Add(pUnit)
	if err != nil {
		// здесь нужно в некоторых случаях делать попытки в цикле
		return err
	}

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
		return nil, models.ErrConnectionCmdInProcessing
	}
	c.startProcessing()
	
	return c.sendBuf, nil
}

func (c *Connection) SendAndReceive(*models.SendBuf) (*models.RecvBuf, error) {
	// отправка данных

	// поллинг

	// очистка буфера отправки

	// получение

	// поллинг
	
	return c.recvBuf, nil
}

func (c *Connection) DrainRecvBuf(*models.RecvBuf) {
	
}


// ------------------------------------------------
// Методы, которые должны вызываться только под захваченным мьютексом

func (c *Connection) isProcessing() bool {
	return c.processing
}

func (c *Connection) startProcessing() {
	c.processing = true
}