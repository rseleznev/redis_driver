package connection

import (
	"context"
	"errors"
	"sync"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
)

// Connector является интерфейсом для внешнего использования
type Connector interface {
	Process(context.Context, []any) (any, error)
	Close()
}

// ------------------------------------------------
// внутренние интерфейсы

type poller interface {
	Add(models.PollingUnit) error
	GetError() error
	DeleteSocketFromPolling(int)
}

type socketer interface {
	GetSocketFd() int
	Connect(*models.Options) error
	Close()
}

type coder interface {
	Encode(*models.SendBuf, []any) error
	Decode([]byte) (any, error)
}

type messenger interface {
	Send([]byte) (int, error)
	Receive(*models.RecvBuf) error
	ChangeSocketFd(int)
}


// ------------------------------------------------

// Connection является одиночным соединением и координирует всю работу от кодирования в RESP
// до возврата результата
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
			Buf: make([]byte, opts.SendBufMinLen),
		},
		recvBuf: &models.RecvBuf{
			Buf: make([]byte, opts.ReceiveBufMinLen),
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
	retriesAvailable := c.opts.RetryAmount
	
	for {
		if retriesAvailable == 0 {
			return models.ErrConnectionRetriesFailed
		}
		
		err := c.socket.Connect(c.opts)
		if err != nil {
			// сокет в неблокирующем режиме и нужно поллить
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EINPROGRESS) {
				err = c.poll("connect")
				if err != nil {
					if err == models.ErrPollTimeout {
						retriesAvailable--

						continue
					}
					
					return err
				}

				return nil
			}

			return err
		}

		break
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
		if ctx.Err() != nil {
			return models.ErrPollTimeout
		}
		
		err = c.poller.Add(pUnit)
		if err != nil {
			// сокет уже поллится, крутимся в цикле и пробуем отдать наше событие
			if err == models.ErrSocketAlreadyAdded {
				continue
			}

			// какие еще ошибки тут могут быть:
			// - асинхронная ошибка
			// - неизвестный тип события
			return c.processPollError(err)
		}
		break
	}

	// блокируемся на чтении результата поллинга
	select {
	case err = <-pUnit.ResultChan:
		if err != nil {
			return c.processPollError(err)
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

// processPollError обрабатывает ошибки poller'а и в случаях, когда что-то можно сделать
// (например, переподключиться), возвращает удобную ошибку вызывающей функции для обработки
func (c *Connection) processPollError(err error) error {

	// флаги ошибок события
	if errors.Is(err, models.ErrSocketEvent) || errors.Is(err, models.ErrSocketHUPEvent) || errors.Is(err, models.ErrSocketRDHUPEvent) {
		return models.ErrConnectionClosed
	}

	// ошибки SO_ERROR
	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ETIMEDOUT) {
		return models.ErrConnectionClosed
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return models.ErrConnectionReset
	}

	return err
}


// ------------------------------------------------
//
// ------------------------------------------------


// Process - основной метод, который выполняет работу
func (c *Connection) Process(ctx context.Context, cmdArgs []any) (any, error) {
	c.mu.Lock()

	if c.isProcessing() {
		c.mu.Unlock()
		
		return nil, models.ErrConnectionCmdInProcess
	}
	c.startProcessing()
	defer c.stopProcessing()

	c.mu.Unlock()
	

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// кодируем в RESP
	err := c.coder.Encode(c.sendBuf, cmdArgs)
	if err != nil {
		return nil, err
	}
	

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// отправляем
	err = c.send(ctx)
	if err != nil {
		return nil, err
	}


	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// получаем
	err = c.receive(ctx)
	if err != nil {
		return nil, err
	}


	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// декодируем
	var result any
	result, err = c.coder.Decode(c.getRecvBufWithWritePos())
	if err != nil {
		return nil, err
	}

	// очищаем буферы
	c.clearBufs()

	return result, nil
}

// send выполняет отправку сообщения
func (c *Connection) send(ctx context.Context) error {
	
	var sentBytes int
	var err error
	retriesAvailable := c.opts.RetryAmount

	for ctx.Err() == nil {
		if retriesAvailable == 0 {
			return models.ErrConnectionRetriesFailed
		}
		
		if c.getSentBytes() == 0 {
			sentBytes, err = c.msgr.Send(c.getSendBuf())
		} else {
			sentBytes, err = c.msgr.Send(c.getSendBufWithoutSentBytes())
		}
		c.addSentBytes(sentBytes)
		if err != nil {
			// буфер отправки полон, нужно ждать
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				err = c.poll("outcome")
				if err != nil {
					// таймаут поллинга истек
					if err == models.ErrPollTimeout {
						retriesAvailable--

						continue
					}
					// соединение закрыто сервером
					if err == models.ErrConnectionReset || err == models.ErrConnectionClosed {
						err = c.reconnect()
						if err != nil {
							return err
						}
						continue
					}
					
					return err
				}

				continue
			}

			switch err {

			// влезли не все данные
			case models.ErrSendMsgTrunc:
				
				continue

			// соединение сброшено
			case models.ErrConnectionReset, models.ErrConnectionClosed:
				err = c.reconnect()
				if err != nil {
					return err
				}
				
				continue

			// сокет не подключен
			case models.ErrNotConnected:
				err = c.connect()
				if err != nil {
					return err
				}
				
				continue

			// все остальные ошибки
			default:
				return err
			
			}	
		}
		break
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	
	return nil
}

// receive выполняет получение сообщения
func (c *Connection) receive(ctx context.Context) error {
	retriesAvailable := c.opts.RetryAmount

	for ctx.Err() == nil {
		if retriesAvailable == 0 {
			return models.ErrConnectionRetriesFailed
		}
		
		err := c.msgr.Receive(c.recvBuf) // передаем структуру, чтобы обработчик указал позицию окончания записи
		if err != nil {
			// нет данных в буфере получения, нужно ждать
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				err = c.poll("income")
				if err != nil {
					// таймаут поллинга истек
					if err == models.ErrPollTimeout {
						retriesAvailable--
						
						continue
					}
					
					return err
				}

				continue
			}

			switch err {

			// в буфер получения влезли не все данные
			case models.ErrRecvMsgTrunc:
				if ok := c.increaseRecvBuf(); ok {
					// увеличили буфер
					continue
				} else {
					// не можем увеличить буфер
					// закрываем соединение и создаем новое
					err = c.reconnect()
					if err != nil {
						return err
					}

					return models.ErrRecvMsgTooBig
				}

			// сокет не подключен
			case models.ErrNotConnected:
				err = c.connect()
				if err != nil {
					return err
				}

				continue

			// все остальные ошибки
			default:
				return err
			
			}
		}
		break
	}
	
	if ctx.Err() != nil {
		return ctx.Err()
	}
	
	return nil
}

func (c *Connection) Close() {
	// закрываем текущий сокет
	c.socket.Close()
}

func (c *Connection) reconnect() error {
	var err error
	
	// закрываем текущий сокет
	c.socket.Close()

	// создаем новый сокет
	c.socket, err = c.factory.NewSocket(c.opts)
	if err != nil {
		return err
	}

	// подключаем новый сокет
	err = c.connect()
	if err != nil {
		return err
	}

	c.msgr.ChangeSocketFd(c.socket.GetSocketFd())

	return nil
}

func (c *Connection) getSendBuf() []byte {
	return c.sendBuf.Buf[:c.getSendBufWritePos()]
}

func (c *Connection) getSendBufWithoutSentBytes() []byte {
	return c.sendBuf.Buf[c.getSentBytes():c.getSendBufWritePos()]
}

func (c *Connection) getSendBufWritePos() int {
	return c.sendBuf.WritePos
}

func (c *Connection) getSentBytes() int {
	return c.sendBuf.SentBytes
}

func (c *Connection) addSentBytes(sentBytes int) {
	c.sendBuf.SentBytes += sentBytes
}

func (c *Connection) increaseRecvBuf() bool {
	// можем ли вообще увеличить буфер
	if c.recvBufFreeSpaceLen() <= 0 {
		return false // увеличить не можем
	}

	// можем ли увеличить в 2 раза
	if c.recvBufLen()*2 <= c.recvBufMaxLen() {
		// увеличиваем в 2 раза
		newBuf := make([]byte, c.recvBufLen()*2)
		copy(newBuf, c.recvBuf.Buf)
		c.recvBuf.Buf = newBuf

		return true
	}

	// увеличиваем насколько возможно
	newBuf := make([]byte, c.recvBufLen()+c.recvBufFreeSpaceLen())
	copy(newBuf, c.recvBuf.Buf)
	c.recvBuf.Buf = newBuf

	return true
}

func (c *Connection) recvBufFreeSpaceLen() int {
	return c.recvBufMaxLen() - c.recvBufLen()
}

func (c *Connection) recvBufMaxLen() int {
	return c.opts.ReceiveBufMaxLen
}

func (c *Connection) recvBufLen() int {
	return len(c.recvBuf.Buf)
}

func (c *Connection) getRecvBufWithWritePos() []byte {
	return c.recvBuf.Buf[:c.getRecvBufWritePos()]
}

func (c *Connection) getRecvBufWritePos() int {
	return c.recvBuf.WritePos
}

// clearBufs очищает буферы.
// Сами данные можем не трогать, т.к. всё завязано на WritePos,
// и поэтому зануляем только WritePos
func (c *Connection) clearBufs() {
	c.sendBuf.WritePos = 0
	c.sendBuf.SentBytes = 0

	c.recvBuf.WritePos = 0
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