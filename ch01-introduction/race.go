package main

// Race Condition Demonstration
// This code demonstrates a race condition where the output is unpredictable.
// The race condition occurs because:
// 1. The main goroutine starts a new goroutine that increments the global 'data' variable
// 2. Both goroutines may access the same memory location concurrently
// 3. There's no synchronization mechanism to ensure atomic operations
// 4. The increment operation (data++) is not atomic - it involves reading, modifying, and writing back
// 5. The timing of when the goroutine runs relative to the main goroutine is non-deterministic
//
// Possible outcomes:
// - If the goroutine hasn't run yet: data = 0, nothing printed
// - If the goroutine has run once: data = 1, nothing printed (condition fails)
// - If the goroutine runs multiple times due to scheduling: data could be 2+, message printed
//
// To detect race conditions, run: go run -race race.go

import (
	"fmt"
	"time"
)

var data int

func increment() {
	// This increment operation is NOT atomic
	// It actually involves three steps:
	// 1. Read current value of data
	// 2. Add 1 to it
	// 3. Write the result back to data
	// During these steps, another goroutine could intervene
	data++
}

func runRaceCondition() {
	// Start a goroutine that will increment the shared variable
	// This creates concurrent access to the global 'data' variable
	go increment()

	// Sleep for 1 second - this is NOT a proper synchronization mechanism!
	// The goroutine might run during this sleep, or it might not
	// The timing is unpredictable and depends on the Go scheduler
	time.Sleep(1 * time.Millisecond)

	// This condition check demonstrates the race condition:
	// We expect data to be 2, but it's actually unpredictable
	// because we have no guarantee about when/how many times
	// the increment function runs
	if data == 0 {
		fmt.Printf("the value is %v.\n", data)
	} else {
		fmt.Printf("Unexpected value: %v (race condition occurred)\n", data)
	}

}

// func main() {
// 	runRaceCondition()
// }

// The Fix
// 1. Use a WaitGroup to wait for the goroutine to finish
// 2. Use a Mutex to synchronize access to the shared variable
// 3. Use a Channel to synchronize access to the shared variable

// we will use atomic operations to fix the race condition next
