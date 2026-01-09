package main

import (
	"fmt"
	"sync"
	"time"
)

// Starvation :- A situation where a concurrent process cannot get the
// resources it needs (locks, CPU time, etc.) to perform its work.
//
// Unlike Livelock, where everyone is stuck equally, Starvation usually
// involves a "greedy" process that unfairly consumes resources at the
// expense of "polite" processes.

var wG sync.WaitGroup // counting semaphore
var sharedLock sync.Mutex

const runtime = 1 * time.Second

func runStarvation() {
	// Greedy worker: Holds the lock for the entire duration of its work.
	// This minimizes the "window of opportunity" for anyone else to grab the lock.
	greedyWorker := func() {
		defer wG.Done()
		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(3 * time.Nanosecond) // Simulated work
			sharedLock.Unlock()
			count++
		}
		fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
	}

	// Polite worker: Only holds the lock for exactly what it needs.
	// It constantly releases and re-acquires, creating many windows
	// where it might lose the lock to the greedy worker.
	politeWorker := func() {
		defer wG.Done()
		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			count++
		}
		fmt.Printf("Polite worker was able to execute %v work loops.\n", count)
	}

	wG.Add(2)
	go politeWorker()
	go greedyWorker()
	wG.Wait()
}

func main() {
	runStarvation()
}

// --- What is happening here? ---
//
// 1. THE RESOURCE: Both workers need the 'sharedLock' to perform their 3ns of work.
//
// 2. THE CRITICAL SECTION:
//    - The Greedy worker expands its critical section to cover all 3ns at once.
//    - The Polite worker breaks its work into three 1ns sections.
//
// 3. THE IMBALANCE:
//    Every time the Polite worker unlocks, it gives the Go scheduler a chance
//    to hand the lock to the Greedy worker. Because the Greedy worker holds the
//    lock longer and more consistently, it effectively "starves" the Polite worker.
//
// 4. THE METRIC:
//    Starvation is identified via metrics. In the output, you will see the
//    Greedy worker completes nearly double the work of the Polite worker in
//    the same 1-second window.
