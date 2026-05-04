package redis_driver

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/options"
)

var (
	longValue, twiceLongValue []byte
	client *Client
	err error
)


func BenchmarkSetShortValue(b *testing.B) {
	ctx := context.Background()
	client, err = NewClient(&options.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,
		SendBufMinLen: 800,
	})
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		err = client.SetValueForKey(ctx, "test1", "value1", time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = err
}

func BenchmarkSetLongValue(b *testing.B) {
	ctx := context.Background()

	longValue, _ = os.ReadFile("./docs/testing/response.json")

	for b.Loop() {
		err = client.SetValueForKey(ctx, "test2", longValue, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = err
}

func BenchmarkSetTwiceLongValue(b *testing.B) {
	ctx := context.Background()

	twiceLongValue = longValue
	twiceLongValue = append(twiceLongValue, longValue...)

	for b.Loop() {
		err = client.SetValueForKey(ctx, "test3", twiceLongValue, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = err
}