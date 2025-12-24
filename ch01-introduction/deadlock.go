package main

import (
	"fmt"
	"sync"
	"time"
)

// Deadlock :- A deadlock is a situation all concurrent operations are waiting for each other to release a resource,
// resulting in a situation where none of the concurrent operations can proceed.
// in this situation the program will never recover without external intervention.

// example of deadlock

type value struct {
	mu sync.Mutex
	v  int
}

var wg sync.WaitGroup

func printSum(v1, v2 *value) {
	defer wg.Done()
	v1.mu.Lock()         // 1
	defer v1.mu.Unlock() // 2

	time.Sleep(2 * time.Second) // 3
	v2.mu.Lock()
	defer v2.mu.Unlock()
	fmt.Printf("sum=%v\n", v1.v+v2.v)
}

var a, b value

func runDeadlock() {
	wg.Add(2)
	go printSum(&a, &b)
	go printSum(&b, &a)
	wg.Wait()
}

func main() {
	runDeadlock()
}

// 1. Here we attempt to enter the critical section for the incoming value.
// 2. Here we use the defer statement to exit the critical section before printSum returns.
// 3. Here we sleep for a period of time to simulate work (and trigger a deadlock).

// when we run this code we should see `fatal error: all goroutines are asleep - deadlock!`

// What is happening here?
// --- Detailed Breakdown of the Deadlock Mechanism ---
//
// 1. THE SETUP:
//    We have two resources (a.mu and b.mu) and two execution paths (Goroutine 1 and 2).
//    - Goroutine 1 calls printSum(&a, &b) -> Wants 'a' then 'b'.
//    - Goroutine 2 calls printSum(&b, &a) -> Wants 'b' then 'a'.
//
// 2. THE TIMING (Race Condition):
//    The time.Sleep() is crucial here. It ensures that both goroutines have enough
//    time to acquire their FIRST lock before either tries to acquire their SECOND lock.
//    Without the sleep, one goroutine might finish the entire function before the
//    other even starts.
//
// 3. THE CIRCULAR WAIT:
//    - Goroutine 1 successfully locks 'a'.
//    - Goroutine 2 successfully locks 'b'.
//    - Goroutine 1 reaches v2.mu.Lock() and blocks, waiting for 'b' to be released.
//    - Goroutine 2 reaches v2.mu.Lock() and blocks, waiting for 'a' to be released.
//    Neither can progress because the 'Unlock' call is deferred and will only
//    trigger after the function completes—which can never happen.
//
// 4. GO RUNTIME DETECTION:
//    Go's runtime is smart enough to see that the main goroutine is blocked on
//    wg.Wait() and all other goroutines are blocked on mutexes. Since no
//    goroutine is left to release a lock, it triggers a fatal panic to prevent
//    the process from hanging forever in a "zombie" state.
//
// --- How to Prevent This ---
//
// A. Lock Ordering: Always acquire locks in the same logical order (e.g., always
//    lock 'a' before 'b' regardless of the function parameters).
// B. Mutex Hierarchy: Assign a rank/ID to each resource and always lock from
//    lowest ID to highest ID.
// C. Use Channels: In Go, it is often better to communicate to share memory
//    rather than sharing memory to communicate (using Mutexes).

// main      printSum(&a,&b)     a.lock      b.lock     printSum(&b,&a)
//    |              |              |           |               |
//    |------------->|              |           |               |
//    |-------------------------------------------------------->|
//    |              |              |           |               |
//    |              |------------->|           |               |
//    |              |              |           |<--------------|
//    |              |              |           |               |
//    |              |      X <-----------------|               |
//    |              |              |----------> X              |
//    |            [DEADLOCK]       |           |           [DEADLOCK]
