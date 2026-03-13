package main

import (
	"fmt"
	"log"

	"github.com/rseleznev/redis_driver/driver"
)

func main() {
	conn, err := driver.NewConn([4]byte{127, 0, 0, 1}, 6379)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	testPing, err := conn.Ping()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(testPing)

	testHello, err := conn.Hello3()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(testHello)

	conn.SetValueForKey("test", "ghdfgdfgsdrrrrfdfgfd", 200)
	v := conn.GetValueByKey("test")

	result, ok := v.([]byte)
	if !ok {
		fmt.Println("ошибка преобразования")
	}
	fmt.Println(string(result))
}