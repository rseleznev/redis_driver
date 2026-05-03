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
	// testData []struct{
	// 	key string
	// 	value any
	// } = []struct{
	// 	key string
	// 	value any
	// }{
	// 	{
	// 		key: "test1",
	// 		value: "value1",
	// 	},
	// 	{
	// 		key: "test2",
	// 		value: []byte{'V', 'A', 'L', 'U', 'E', '2'},
	// 	},
	// 	{
	// 		key: "test3",
	// 		value: longValue,
	// 	},
	// 	{
	// 		key: "test4",
	// 		value: twiceLongValue,
	// 	},
	// }
)

func BenchmarkSet(b *testing.B) {
	ctx := context.Background()
	client, err = NewClient(&options.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,
		SendBufMinLen: 800,
	})

	if err != nil {
		b.Fatal(err)
	}

	longValue, _ = os.ReadFile("./docs/testing/response.json")
	twiceLongValue = longValue
	twiceLongValue = append(twiceLongValue, longValue...)

	for b.Loop() {
		err = client.SetValueForKey(ctx, "test", twiceLongValue, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = err
}