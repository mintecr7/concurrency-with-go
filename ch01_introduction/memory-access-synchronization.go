package main

import (
	"fmt"
	"sync"
	"time"
)

func memoryAccessSynchronization() {
	var memoryAccess sync.Mutex // 1
	var data int
	go func() {
		// time.Sleep(5 * time.Millisecond) // race condition will occur in this case [print statement is called first]
		memoryAccess.Lock()   // 2
		data++                // atomic
		memoryAccess.Unlock() // 3
	}()
	time.Sleep(10 * time.Millisecond) // prevent race condition, but not a good solution
	memoryAccess.Lock()               // 4
	if data == 0 {
		fmt.Printf("the value is %v.\n", data) // atomic
	} else {
		fmt.Printf("the value is %v.\n", data) // atomic
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

// This example might have solved the data race,but we haven't solved our race condition.
// The order of operations in this program is still nondeterministic;
// we’ve just narrowed the scope of the nondeterminism a bit. In this example,
// either the goroutine will execute first, or both our if and else blocks will.
// We still don't know which will occur first in any given execution of this program.

// synchronizing access to the memory might cause problems like deadlock, starvation, and live-lock.
