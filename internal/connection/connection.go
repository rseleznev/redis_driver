package connection

import (
	"context"
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
	DeleteSocketFromPolling(int)
}

type sender interface {
	Send([]byte) (int, error)
}

type receiver interface {
	Receive([]byte) error
}

type encoder interface {
	Encode([]byte, []any) ([]byte, error)
}

type decoder interface {
	Decode([]byte) (any, error)
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
	enc encoder
	dec decoder

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
		Buf: make([]byte, 0, opts.SendBufAvgLen),
		PrevCap: opts.SendBufAvgLen,
	}
	conn.recvBuf = &models.RecvBuf{
		SocketFd: s.GetSocketFd(),
		Buf: make([]byte, 0, opts.ReceiveBufAvgLen),
		PrevCap: opts.ReceiveBufAvgLen,
	}

	conn.sender = f.NewSender()
	conn.receiver = f.NewReceiver()
	
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
	ctx, cancel := c.newContextWithTimeout()
	defer cancel()

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
	select {
	case err = <-pUnit.ResultChan:
		if err != nil {
			return c.processPollError(eventType, err)
		}

	case <-ctx.Done():
		c.poller.DeleteSocketFromPolling(c.socket.GetSocketFd())

		return models.ErrPollTimeout

	}
	

	return nil
}

func (c *Connection) newContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.opts.PollingTimeout)
}

// processPollError обрабатывает ошибки epoll в зависимости от события.
// Если с ошибкой что-то можно сделать, это будет сделано.
// Например, может переподключиться к серверу или делать ретраи отправки/получения
func (c *Connection) processPollError(_ string, _ error) error {

	// внимательно изучить, в каких ситуациях какие флаги устанавливает epoll
	
	return nil
}

func (c *Connection) Process(cmdParams []any) (any, error) {
	c.mu.Lock()
	
	defer c.mu.Unlock()

	if c.isProcessing() {
		return nil, models.ErrConnectionCmdInProcess
	}
	c.startProcessing()
	
	// кодируем в RESP
	c.enc.Encode(c.sendBuf.Buf, cmdParams)

	err := c.send(0)
	if err != nil {
		return nil, err
	}

	err = c.receive()
	if err != nil {
		return nil, err
	}

	// декодируем в объект Go
	c.dec.Decode(c.recvBuf.Buf)

	return nil, nil
}

// CancelProcessing отменяет начатую операцию
func (c *Connection) CancelProcessing() {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.drainRecvBuf()
	c.stopProcessing()
}

func (c *Connection) send(fromIdx int) error {
	var sentBytes int
	var err error
	
	if fromIdx == 0 {
		sentBytes, err = c.sender.Send(c.sendBuf.Buf)
	} else {
		sentBytes, err = c.sender.Send(c.sendBuf.Buf[fromIdx:])
	}
	if err != nil {
		// буфер отправки полон, нужно ждать
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			err = c.poll("outcome")
			if err != nil {
				return err
			}

			return nil
		}

		switch err {

		// влезли не все данные
		case models.ErrSendMsgTrunc:
			return c.send(sentBytes)

		// соединение сброшено
		case models.ErrConnectionReset, models.ErrConnectionClosed:
			// переподключаемся

		// сокет не подключен
		case models.ErrNotConnected:
			// коннектимся

		// все остальные ошибки
		default:
			return err
		
		}	
	}
	
	return nil
}

func (c *Connection) receive() error {
	err := c.receiver.Receive(c.recvBuf.Buf)
	if err != nil {
		// нет данных в буфере получения, нужно ждать
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			err = c.poll("income")
			if err != nil {
				return err
			}

			return c.receive()
		}

		switch err {

		// в буфер получения влезли не все данные
		case models.ErrRecvMsgTrunc:
			// увеличиваем буфер?

		// соединение сброшено
		case models.ErrConnectionClosed:
			// переподключаемся

		// сокет не подключен
		case models.ErrNotConnected:
			// коннектимся

		// все остальные ошибки
		default:
			return err
		
		}
	}
	
	return nil
}

// Done очищает буфер получения, скидывает флаг процессинга.
// Вызывается, когда команда прочитала все свои данные
func (c *Connection) Done() {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.stopProcessing()
	c.drainRecvBuf()
	c.correctSendBufCap()
	c.correctReceiveBufCap()
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

func (c *Connection) correctSendBufCap() {
	// сравниваем текущую емкость с SendBuf.PrevCap
}

func (c *Connection) correctReceiveBufCap() {

}