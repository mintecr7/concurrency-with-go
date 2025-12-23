package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Atomicity :- When something is atomic, it means that it is an indivisible/uninterruptible
// operation with in a context.
// Something may be atomic in one context, but not another.
// In other words, the atomicity of an operation can change depending on the currently defined scope.

// When thinking about atomicity, very often the first thing you need to do is to define
// the context, or scope, the operation will be considered to be atomic in.

// Indivisibility :- an operation that proceed in its entirety, without anything happening in that context simultaneously.
// e.g: i++ not atomic because it involves multiple steps (read, modify, write) and can be interrupted.
// e.g: i++ atomic in the scope of a function.

// BAD EXAMPLE: Non-atomic operation leading to race condition
func badAtomicityExample() {
	var counter int

	// Start multiple goroutines that increment the same variable
	for i := 0; i < 1000; i++ {
		go func() {
			counter++ // NOT atomic! Read-modify-write operation
		}()
	}

	// This will likely print a value less than 1000 due to race conditions
	fmt.Printf("Bad example result: %d (expected: 1000)\n", counter)
}

// GOOD EXAMPLE 1: Using mutex for atomicity
func goodAtomicityWithMutex() {
	var counter int
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock() // Ensure exclusive access
			counter++ // Now atomic within the critical section
			mu.Unlock()
		}()
	}
	wg.Wait()

	fmt.Printf("Good example (mutex) result: %d (expected: 1000)\n", counter)
}

// GOOD EXAMPLE 2: Using atomic operations
func goodAtomicityWithAtomic() {
	var counter int64

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1) // Atomic operation
		}()
	}
	wg.Wait()

	fmt.Printf("Good example (atomic) result: %d (expected: 1000)\n", counter)
}

func demoAtomicityExamples() {
	fmt.Println("=== Atomicity Examples ===")

	fmt.Println("\n1. Bad Example (Non-atomic):")
	badAtomicityExample()

	fmt.Println("\n2. Good Example (Mutex):")
	goodAtomicityWithMutex()

	fmt.Println("\n3. Good Example (Atomic Operations):")
	goodAtomicityWithAtomic()

}

// To run this demo, create a main function that calls demoAtomicityExamples()
// or run this file separately: go run atomicity.go
// func main() {
// 	demoAtomicityExamples()
// }
