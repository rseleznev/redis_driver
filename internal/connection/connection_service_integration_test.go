package connection

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
)

var opts = &models.Options{
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
var conn Connector
var connErr error


func TestMain(m *testing.M) {
	conn, connErr = NewConnector(opts)
	if connErr != nil {
		panic(connErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	
	r, err := conn.Process(ctx, []any{"HELLO", "3"})
	if err != nil {
		panic(err)
	}
	cancel()

	mapResult, ok := r.(map[string]string)
	if !ok {
		panic("ошибка утверждения типа")
	}

	fmt.Println("Вывод команды HELLO 3:", mapResult)

	code := m.Run()

	conn.Close()
	os.Exit(code)
}

func TestShortString(t *testing.T) {
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	
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

	t.Log("Вывод команды TestShortString.SET:", string(res))

	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)

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

	t.Log("Вывод команды TestShortString.GET:", string(res))
}

func TestShortBytes(t *testing.T) {
	value := []byte{'B', 'Y', 'T', 'E', 'S'}
	
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	
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

	t.Log("Вывод команды TestShortBytes.SET:", string(res))


	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)

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

	t.Log("Вывод команды TestShortBytes.GET:", string(res))
}

func TestLongValue(t *testing.T) {
	longValue, _ := os.ReadFile("./../../docs/testing/response.json")
	t.Logf("Длина исходной строки: %d \n", len(longValue))
	
	// Записываем ключ
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	
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

	t.Log("Вывод команды TestLongValue.SET:", string(res))


	// Читаем ключ
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)

	r, err = conn.Process(ctx, []any{"GET", "longValue"})
	if err != nil {
		t.Error(err)
	}
	cancel()

	res, ok = r.([]byte)
	if !ok {
		t.Error("Ошибка утверждения типа")
	}
	t.Logf("Длина полученной строки: %d \n", len(res))

	if slices.Compare(res, longValue) != 0 {
		t.Errorf("Ожидаемый результат %s, получено %s", longValue, res)
	}

	t.Log("Вывод команды TestLongValue.GET:", string(res))
}