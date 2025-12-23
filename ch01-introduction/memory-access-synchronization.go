package main

import (
	"fmt"
	"sync"
)

func memoryAccessSynchronization() {
	var memoryAccess sync.Mutex // 1
	var data int
	go func() {
		memoryAccess.Lock() // 2
		data++
		memoryAccess.Unlock() // 3
	}()
	memoryAccess.Lock() // 4
	if data == 0 {
		fmt.Printf("the value is %v.\n", data)
	} else {
		fmt.Printf("the value is %v.\n", data)
	}
	memoryAccess.Unlock() // 5
}

// func main() {
// 	memoryAccessSynchronization()
// }

// 1. Here we add a variable that will allow our code to synchronize access to the data variable’s memory.
// 2. Here we declare that until we declare otherwise, our goroutine should have
// exclusive access to this memory.
// 3. Here we declare that the goroutine is done with this memory.
// 4. Here we once again declare that the following conditional statements should have
// exclusive access to the data variable’s memory.
// 5. Here we declare we’re once again done with this memory.
