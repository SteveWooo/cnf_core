package main

import (
	"fmt"
)

func test() {
	fmt.Println("do test")
	for {

	}
}

func main() {
	chanel := make(chan int, 5)

	for {
		chanel <- 1
		go test()
	}
}