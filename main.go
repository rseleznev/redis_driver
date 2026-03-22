package main

import (
	"fmt"
	"log"

	"github.com/rseleznev/redis_driver/driver"
	"github.com/rseleznev/redis_driver/internal/models"
)

func main() {
	conn, err := driver.NewConn(models.Options{
		RedisIp: [4]byte{127, 0, 0, 1},
		RedisPort: 6379,
		RetryAmount: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// Ping
	testPing, err := conn.Ping()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(testPing)

	// Hello
	err = conn.Hello3()
	if err != nil {
		fmt.Println(err)
	}

	// Set
	conn.SetValueForKey("test", "value", 200)
	// Get
	v := conn.GetValueByKey("d41d8cd98f00b204e9800998ecf8427e")

	result, ok := v.([]byte)
	if !ok {
		fmt.Println("ошибка преобразования")
	}
	fmt.Println(string(result))
}