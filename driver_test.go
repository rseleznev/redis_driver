package redis_driver

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var (
	conn *Conn
) 

func TestMain(m *testing.M) {
	var err error

	conn, err = NewConn(models.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,
		RetryAmount: 3,
	})
	if err != nil {
		fmt.Println("Ошибка соединения: ", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	conn.Close()
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
	
	data := []struct {
		testName string
		key string
		value any
		duration int
		expectedErr error
	}{
		{"success", "test", "value", 300, nil},
		{"successPermKey", "testPermKey", "value2", 0, nil},
	}

	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			err := conn.SetValueForKey(d.key, d.value, d.duration)
			if d.expectedErr != err {
				t.Fatal("Ошибка в команде SetValueForKey: ", err)
			}
		})
	}
	t.Log("Выполнена команда SetValueForKey")
}

func TestGetValueByKey(t *testing.T) {
	t.Log("Запуск команды GetValueByKey")

	data := []struct {
		testName string
		key string
		expectedStringValue string
		expectedBytesValue []byte
		expectedErr error
	}{
		{"success", "test", "value", nil, nil},
		{"successPermKey", "testPermKey", "value2", nil, nil},
	}

	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			r, err := conn.GetValueByKey(d.key)
			if d.expectedErr != err {
				t.Fatalf("Ожидаемая ошибка %s, получено %s", d.expectedErr, err)
			}
			
			switch result := r.(type) {
			case string:
				if d.expectedStringValue != result {
					t.Fatalf("Ожидаемое значение %s, получено %s", d.expectedStringValue, result)
				}
			
			case []byte:
				if slices.Compare(d.expectedBytesValue, result) == 0 {
					t.Fatalf("Ожидаемое значение %s, получено %s", d.expectedBytesValue, result)
				}

			}
		})
	}
	t.Log("Выполнена команда GetValueByKey")
}