package main

import (
	"fmt"
	"log"
	// "sync"

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

	conn.GetValueByKey("d41d8cd98f00b204e9800998ecf8427e")

	// var wg sync.WaitGroup

	// wg.Add(2)

	// for range 2 {
	// 	go func() {
	// 		test, err := conn.Ping()
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		fmt.Println("Горутина main:", test)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()
}