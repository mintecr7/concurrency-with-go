package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// TOPIC: The Difference Between Concurrency and Parallelism
// =============================================================================
// - Concurrency is a property of the code.
//		- It refers to the design and structure of the program.
// 		- It is the intent that tasks can run independently.
// - Parallelism is a property of the running program.
//		- It refers to the actual execution state.
// 		- It depends on the hardware (e.g., does the machine have multiple cores?).
//	Key Insight: You do not write "parallel code." You write concurrent code and hope that the runtime and hardware allow it to execute in parallel.
// CORE DEFINITION:
// "Concurrency is a property of the code; parallelism is a property of the running program."
//
// 1. The Distinction
//    - Concurrency: The design/intent. Splitting a program into chunks that *can*
//      run independently. We write concurrent code.
//    - Parallelism: The execution. Whether those chunks actually run at the
//      exact same instant. We hope for parallel execution.
//
// 2. The Illusion of Parallelism
//    - On a single-core machine, concurrent code executes sequentially but
//      context-switches so fast it looks parallel.
//    - On a multi-core machine, the code may actually run in parallel.
//    - KEY POINT: It is desirable to be ignorant of this. We rely on abstraction
//      layers (Runtime -> OS -> CPU) to manage the execution.
//
// 3. Context & Relativity
//    - Parallelism is a function of the context (bounds).
//    - Time Context: Two 1-second tasks running inside a 5-second window are "parallel"
//      relative to that window. Inside a 1-second window, they are sequential.
//    - System Context: Processes are isolated by the OS. Threads share memory.
//
// 4. The Hierarchy of Difficulty
//    - Process Level: High isolation, easier logic (but resource conflicts exist).
//    - Thread Level: Low isolation, shared memory. Hardest to get right (race conditions, deadlocks, livelock and starvation).
//    - Go Level: Goroutines supplant threads. They bring the ease of concurrent reasoning
//      back to the developer without the overhead of OS threads.

// =============================================================================
// DEMONSTRATION
// =============================================================================

// SimulateWork represents a "chunk" of our concurrent code.
func SimulateWork(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Task %d starting\n", id)
	time.Sleep(100 * time.Millisecond) // Simulating CPU work
}

// CPUBoundWork does actual CPU-intensive work (not just sleeping)
// This forces the runtime to actually use CPU cores
func CPUBoundWork(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	// CPU-intensive calculation
	sum := 0
	for i := range 1000000000 {
		sum += i
	}

	elapsed := time.Since(start)
	fmt.Printf("Task %d finished in %v (sum=%d)\n", id, elapsed, sum)
}

// RunDemo proves that Parallelism is a property of the Runtime, not the code.
// We run the EXACT same code twice, but change the Runtime context.
func RunDemo() {
	numTasks := 4

	// Scenario 1: Single Core - Tasks run sequentially
	fmt.Println("=== SINGLE CORE (Concurrent but NOT Parallel) ===")
	runtime.GOMAXPROCS(1)

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(numTasks)
	for i := 1; i <= numTasks; i++ {
		go CPUBoundWork(i, &wg)
	}
	wg.Wait()
	elapsed1 := time.Since(start)
	fmt.Printf("Total time (1 core): %v\n\n", elapsed1)

	// Scenario 2: Multiple Cores - Tasks can run in parallel
	fmt.Println("=== MULTIPLE CORES (Concurrent AND Parallel) ===")
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Using %d cores\n", runtime.NumCPU())

	start = time.Now()
	wg.Add(numTasks)
	for i := 5; i <= numTasks+4; i++ {
		go CPUBoundWork(i, &wg)
	}
	wg.Wait()
	elapsed2 := time.Since(start)
	fmt.Printf("Total time (%d cores): %v\n\n", runtime.NumCPU(), elapsed2)

	// Show the speedup
	speedup := float64(elapsed1) / float64(elapsed2)
	fmt.Printf("=== RESULTS ===\n")
	fmt.Printf("Speedup: %.2fx faster with multiple cores\n", speedup)
	fmt.Printf("This proves parallelism! Same code, different execution.\n")
}
