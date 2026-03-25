package redis_driver

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var conn *Conn

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
		t.Error("Ошибка в команде Ping: ", err)
	}
	t.Log("Результат команды Ping: ", testPing)
}

func TestHello3(t *testing.T) {
	t.Parallel()
	t.Log("Запуск команды Hello3")
	
	err := conn.Hello3()
	if err != nil {
		t.Error("Ошибка в команде Hello3: ", err)
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
		{"successBytes", "testBytes", []byte{'h', 'e', 'l', 'l', 'o'}, 300, nil},
		{"failWithMap", "testFail", map[string]string{"tt": "vv"}, 300, models.ErrUnsupportedDataType},
	}

	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			err := conn.SetValueForKey(d.key, d.value, d.duration)
			if d.expectedErr != err {
				t.Errorf("Ожидаемая ошибка %s, получено %s", d.expectedErr, err)
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
		expectedValueType string
		expectedValue any
		expectedErr error
	}{
		{"success", "test", "string", "value", nil},
		{"successPermKey", "testPermKey", "string", "value2", nil},
		{"successBytes", "testBytes", "bytes", []byte{'h', 'e', 'l', 'l', 'o'}, nil},
		{"failNoValue", "testNoValue", "nil", nil, models.ErrNoValue},
	}

	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			r, err := conn.GetValueByKey(d.key)
			if d.expectedErr != err {
				t.Errorf("Ожидаемая ошибка %s, получено %s", d.expectedErr, err)
			}

			switch d.expectedValueType {
			case "string":
				resultBytes, ok := r.([]byte)
				if !ok {
					t.Fatal("Ошибка преобразования в байты")
				}
				result := string(resultBytes)
				if d.expectedValue != result {
					t.Errorf("Ожидаемое значение %s, получено %s", d.expectedValue, result)
				}

			case "bytes":
				result, ok := r.([]byte)
				if !ok {
					t.Error("Ошибка преобразования в байты")
				}
				expSlice := d.expectedValue.([]byte)
				if slices.Compare(expSlice, result) != 0 {
					t.Errorf("Ожидаемое значение %s, получено %s", expSlice, result)
				}

			case "nil":
				if d.expectedValue != r {
					t.Errorf("Ожидаемое значение %s, получено %s", d.expectedValue, r)
				}

			}
		})
	}
	t.Log("Выполнена команда GetValueByKey")
}

func TestProtocolErrorAndReconnect(t *testing.T) {
	cmd := []byte{
		'*', '2', '\r', '\n',
		'$', '3', '\r', '\n',
		'S', 'E', 'T', '\r', '\n',
		'%', '1', '\r', '\n',
		'$', '4', '\r', '\n',
		't', 'e', 's', 't', '\r', '\n',
		'$', '2', '\r', '\n',
		'v', 'h', '\r', '\n',
	}

	err := conn.incorrectTestCommand(cmd)
	if !errors.Is(err, models.ErrRedisProtocol) {
		t.Error("Некорректная ошибка в TestProtocolErrorAndReconnect")
	}

	err = conn.SetValueForKey("testAfterReconn", "valueAfterReconn", 300)
	if err != nil {
		t.Error("Неожиданная ошибка: ", err)
	}

	r, err := conn.GetValueByKey("testAfterReconn")
	if err != nil {
		t.Error("Неожиданная ошибка: ", err)
	}

	resultBytes, ok := r.([]byte)
	if !ok {
		t.Error("Ошибка преобразования в байты")
	}
	result := string(resultBytes)
	if result != "valueAfterReconn" {
		t.Error("Значение не соответствует ожидаемому. Значение: ", result)
	}
}