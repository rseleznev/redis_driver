package redis_driver

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/options"
)

var (
	testClient = &Client{}
	commandTimeout time.Duration = time.Millisecond*100
)

func TestMain(m *testing.M) {
	client, err := NewClient(&options.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,

		SetKeepAlive: true,

		SendBufMinLen: 256,
		SendBufMaxLen: 1024,
		ReceiveBufMinLen: 256,
		ReceiveBufMaxLen: 1024,
	})
	if err != nil {
		panic(err)
	}

	testClient = client

	code := m.Run()
	testClient.Close()
	os.Exit(code)
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	result, err := testClient.Ping(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	if result != "PONG" {
		t.Fatal("Неожиданный результат")
	}
	t.Log(result)
}

func TestHello3(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	result, err := testClient.Hello3(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	t.Log(result)
}

func TestSet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	err := testClient.SetValueForKey(ctx, "func_test", "value", time.Minute*2)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
}

func TestGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	
	result, err := testClient.GetValueByKey(ctx, "func_test")
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	t.Log(string(result))
}