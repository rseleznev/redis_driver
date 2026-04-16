package connection

import (
	"context"
	"errors"
	"sync"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
)

type Connector interface {
	Process(context.Context, []any) (any, error)
}


type poller interface {
	Add(models.PollingUnit) error
	GetError() error
	DeleteSocketFromPolling(int)
}

type socketer interface {
	GetSocketFd() int
	Connect(*models.Options) error
}

type coder interface {
	Encode(*models.SendBuf, []any) error
	Decode([]byte) (any, error)
}

type messenger interface {
	Send([]byte) (int, error)
	Receive(*models.RecvBuf) error
}


type Connection struct {
	opts *models.Options
	mu sync.Mutex

	// фабрика для создания нужных объектов, 
	// сохраняем в структуру для использования в будущем
	factory Factory

	// интерфейсы
	poller poller
	socket socketer
	coder coder
	msgr messenger

	// буферы
	sendBuf *models.SendBuf
	recvBuf *models.RecvBuf

	// флаг, занято ли соединение
	processing bool
}

func NewConnector(opts *models.Options) (Connector, error) {
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
		opts: opts,
		mu: sync.Mutex{},
		factory: f,
		poller: p,
		socket: s,

		sendBuf: &models.SendBuf{
			Buf: make([]byte, 0, opts.SendBufMinLen),
		},
		recvBuf: &models.RecvBuf{
			Buf: make([]byte, 0, opts.ReceiveBufMinLen),
		},
	}

	// подключаем сокет
	err = conn.connect()
	if err != nil {
		return nil, err
	}

	conn.coder = f.NewCoder()
	conn.msgr = f.NewMessenger(conn.socket.GetSocketFd())
	
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
		SocketFd: c.socket.GetSocketFd(),
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


// ------------------------------------------------

// ------------------------------------------------


func (c *Connection) Process(ctx context.Context, cmdArgs []any) (any, error) {
	c.mu.Lock()
	
	defer c.mu.Unlock()

	if c.isProcessing() {
		return nil, models.ErrConnectionCmdInProcess
	}
	c.startProcessing()
	
	// кодируем в RESP
	err := c.coder.Encode(c.sendBuf, cmdArgs)
	if err != nil {
		// обработка увеличения буфера отправки!

		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	
	// отправляем
	err = c.send(0)
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// получаем
	err = c.receive()
	if err != nil {
		// обработка увеличение буфера получения!	

		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// декодируем
	var result any
	result, err = c.coder.Decode(c.recvBuf.Buf[:c.recvBuf.WritePos])

	// очищаем буферы?

	c.stopProcessing()

	return result, nil
}

func (c *Connection) send(fromIdx int) error {
	var sentBytes int
	var err error
	
	if fromIdx == 0 {
		sentBytes, err = c.msgr.Send(c.sendBuf.Buf)
	} else {
		sentBytes, err = c.msgr.Send(c.sendBuf.Buf[fromIdx:])
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
	err := c.msgr.Receive(c.recvBuf)
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