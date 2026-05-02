package connection

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
	"github.com/rseleznev/redis_driver/options"
)

var opts = &options.Options{
	RedisIp: [4]byte{127, 0, 0, 1},
	RedisPort: 6379,
	RetryAmount: 3,

	SetKeepAlive: true,
	KeepAliveIdle: 300,
	KeepAliveInterval: 60,
	KeepAliveCheckAmount: 5,

	SendBufMinLen: 256,
	SendBufMaxLen: 1024,
	ReceiveBufMinLen: 256,
	ReceiveBufMaxLen: 1024,

	PollingTimeout: time.Millisecond*50,
}
var conn *Connection
var commandTimeout time.Duration = time.Millisecond*100


func TestMain(m *testing.M) {
	// Создаем Connection, чтобы иметь возможность мокать отдельные интерфейсы
	var f Factory
	
	// создаем сокет
	s, err := f.NewSocket(opts)
	if err != nil {
		panic(err)
	}

	// создаем поллер
	p, err := f.NewPoller()
	if err != nil {
		panic(err)
	}

	conn = &Connection{
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
		panic(err)
	}

	conn.coder = f.NewCoder()
	conn.msgr = f.NewMessenger(conn.socket.GetSocketFd())

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	r, err := conn.Process(ctx, []any{"HELLO", "3"})
	if err != nil {
		panic(err)
	}
	cancel()

	mapResult, ok := r.(map[string]string)
	if !ok {
		panic("ошибка утверждения типа")
	}

	fmt.Println("Вывод команды HELLO 3: ", mapResult)

	code := m.Run()

	conn.Close()
	os.Exit(code)
}

func TestShortString(t *testing.T) {
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	r, err := conn.Process(ctx, []any{"SET", "shortString", "short string value", "PX", "50000"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok := r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if string(res) != "OK" {
		t.Error("Некорректный результат")
	}

	t.Log("Вывод команды TestShortString.SET: ", string(res))

	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), commandTimeout)

	r, err = conn.Process(ctx, []any{"GET", "shortString"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok = r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if string(res) != "short string value" {
		t.Error("Некорректный результат")
	}

	t.Log("Вывод команды TestShortString.GET: ", string(res))
}

func TestShortBytes(t *testing.T) {
	value := []byte{'B', 'Y', 'T', 'E', 'S'}
	
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	r, err := conn.Process(ctx, []any{"SET", "shortBytes", value, "PX", "50000"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok := r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if string(res) != "OK" {
		t.Error("Некорректный результат")
	}

	t.Log("Вывод команды TestShortBytes.SET: ", string(res))


	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), commandTimeout)

	r, err = conn.Process(ctx, []any{"GET", "shortBytes"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok = r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if slices.Compare(res, value) != 0 {
		t.Errorf("Ожидаемый результат %s, получено %s", value, res)
	}

	t.Log("Вывод команды TestShortBytes.GET: ", string(res))
}

func TestLongValue(t *testing.T) {
	longValue, _ := os.ReadFile("./../../docs/testing/response.json")
	
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	r, err := conn.Process(ctx, []any{"SET", "longValue", longValue, "PX", "50000"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok := r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if string(res) != "OK" {
		t.Error("Некорректный результат")
	}

	t.Log("Вывод команды TestLongValue.SET: ", string(res))


	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), commandTimeout)

	r, err = conn.Process(ctx, []any{"GET", "longValue"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok = r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	if slices.Compare(res, longValue) != 0 {
		t.Errorf("Ожидаемый результат %s, получено %s", longValue, res)
	}

	t.Log("Вывод команды TestLongValue.GET: ", string(res))
}

func TestWithReconnect(t *testing.T) {
	conn.coder = mockCoder{ // mockCoder из файла connection_test.go
		encodeFunc: func(sb *models.SendBuf, a []any) error {
			data := []byte{
				'*', '1', '\r', '\n',
				'%', '1', '\r', '\n',
				'$', '4', '\r', '\n',
				'T', 'E', 'S', 'T', '\r', '\n',
				'$', '2', '\r', '\n',
				'O', 'K', '\r', '\n',
			}

			n := copy(sb.Buf, data)
			sb.WritePos = n

			return nil
		},
		decodeFunc: func(b []byte) (any, error) {
			var idx int = 1
			errValue := make([]byte, 0, 50)

			for {
				if b[idx] == '\r' {
					idx++

					if b[idx] == '\n' {
						break
					}
				}

				errValue = append(errValue, b[idx])
				idx++
			}
			return nil, fmt.Errorf("%s: %w", errValue, models.ErrRedisException)
		},
	}


	// Отправляем некорректный запрос (мапу) намеренно, чтобы получить ошибку протокола и сброс соединения
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	_, err := conn.Process(ctx, []any{"wrongDataType"})
	if err != nil {
		t.Logf("Ошибка протокола (ожидаемая): %s", err)
	}
	cancel()

	// Отправляем нормальную команду, которая должна быть обработана уже другим сокетом
	conn.coder = conn.factory.NewCoder()
	ctx, cancel = context.WithTimeout(context.Background(), commandTimeout)
	
	r, err := conn.Process(ctx, []any{"HELLO", "3"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok := r.(map[string]string)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}

	t.Log("Вывод команды TestWithReconnect.HELLO 3: ", res)
}

func TestMultipleStreams(t *testing.T) {
	var wg sync.WaitGroup

	// первая горутина вызывает команду PING
	wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)

		var r any
		var err error
	
		for ctx.Err() == nil {
			r, err = conn.Process(ctx, []any{"PING"})
			if err != nil {
				if err == models.ErrConnectionCmdInProcess {
					t.Log("Соединение занято")
					continue
				}
				t.Error(err)
			}
			break
		}
		cancel()

		res, ok := r.([]byte)
		if !ok {
			t.Error("Ошибка утверждения типа")
		}

		t.Log("Вывод команды TestWithReconnect.PING: ", string(res))
	})

	// вторая горутина хочет получить значение по ключу multiple_streams_test
	// (которое будет указано третьей горутиной)
	wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)

		var r any
		var err error
	
		for ctx.Err() == nil {
			r, err = conn.Process(ctx, []any{"GET", "multiple_streams_test"})
			if err != nil {
				if err == models.ErrConnectionCmdInProcess {
					t.Log("Соединение занято")
					continue
				}
				if err == models.ErrNoValue {
					t.Log("Еще нет значения")
					continue
				}
				t.Error(err)
			}
			break
		}
		cancel()

		res, ok := r.([]byte)
		if !ok {
			t.Error("Ошибка утверждения типа")
		}

		t.Log("Вывод команды TestWithReconnect.GET: ", string(res))
	})

	// третья горутина указывает значение ключа multiple_streams_test
	wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)

		var r any
		var err error
	
		for ctx.Err() == nil {
			r, err = conn.Process(ctx, []any{"SET", "multiple_streams_test" , "testValue", "PX", "1000"})
			if err != nil {
				if err == models.ErrConnectionCmdInProcess {
					t.Log("Соединение занято")
					continue
				}
				t.Error(err)
			}
			break
		}
		cancel()

		res, ok := r.([]byte)
		if !ok {
			t.Error("Ошибка утверждения типа")
		}

		t.Log("Вывод команды TestWithReconnect.SET: ", string(res))
	})

	wg.Wait()
}