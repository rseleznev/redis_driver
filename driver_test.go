package redis_driver

import (
	"os"
	"fmt"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var (
	conn *Conn
	err error
	key = "test"
	value = "value"
	duration = 300
) 

func TestMain(m *testing.M) {
	conn, err = NewConn(models.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,
		RetryAmount: 3,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestPing(t *testing.T) {
	t.Parallel()
	t.Log("Запуск команды Ping")
	
	testPing, err := conn.Ping()
	if err != nil {
		t.Fatal("Ошибка в команде Ping: ", err)
	}
	t.Log("Результат команды Ping: ", testPing)
}

func TestHello3(t *testing.T) {
	t.Parallel()
	t.Log("Запуск команды Hello3")
	
	err := conn.Hello3()
	if err != nil {
		t.Fatal("Ошибка в команде Hello3: ", err)
	}
	t.Log("Выполнена команда Hello3")
}

func TestSetValueForKey(t *testing.T) {
	t.Log("Запуск команды SetValueForKey")
	
	err := conn.SetValueForKey(key, value, duration)
	if err != nil {
		t.Fatal("Ошибка в команде SetValueForKey: ", err)
	}
	t.Log("Выполнена команда SetValueForKey")
}

func TestGetValueByKey(t *testing.T) {
	t.Log("Запуск команды GetValueByKey")
	
	testGet := conn.GetValueByKey(key)

	var result string

	switch r := testGet.(type) {
	case string:
		result = r

	case []byte:
		result = string(r)

	default:
		t.Fatal("Неожиданный тип у значения: ", testGet)

	}
	if result != value {
		t.Fatalf("Значение не совпадает - ожидаемое %s, получено %s", value, result)
	}

	t.Log("Результат команды GetValueByKey: ", result)
}