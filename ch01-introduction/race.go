package main

import (
	"fmt"
	"time"
)

var data int

func increment() {
	data++
}

func main() {
	go increment()
	time.Sleep(1 * time.Second)

	if data == 2 {
		fmt.Printf("the value is %v.\n", data)
	}

}
