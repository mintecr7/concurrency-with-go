package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// 1. BASIC GOROUTINE CREATION
// ============================================================================

// basicGoroutine demonstrates the simplest way to create a goroutine
func basicGoroutine() {
	fmt.Println("\n=== Basic Goroutine ===")

	// Every Go program has at least one goroutine: the main goroutine
	// We can create more using the 'go' keyword

	go sayHello() // Creates a new goroutine running sayHello()

	// Problem: Without synchronization, main might exit before sayHello runs!
	time.Sleep(100 * time.Millisecond) // BAD: Creates a race condition
}

func sayHello() {
	fmt.Println("hello from goroutine")
}

// ============================================================================
// 2. ANONYMOUS FUNCTIONS AS GOROUTINES
// ============================================================================

func anonymousGoroutines() {
	fmt.Println("\n=== Anonymous Function Goroutines ===")

	// Method 1: Inline anonymous function
	go func() {
		fmt.Println("hello from inline anonymous goroutine")
	}() // Notice the () to invoke immediately

	// Method 2: Assign to variable first
	sayWelcome := func() {
		fmt.Println("welcome from variable goroutine")
	}
	go sayWelcome()

	time.Sleep(100 * time.Millisecond) // Still a race condition!
}

// ============================================================================
// 3. JOIN POINTS WITH sync.WaitGroup
// ============================================================================

// properSynchronization shows the CORRECT way to wait for goroutines
func properSynchronization() {
	fmt.Println("\n=== Proper Synchronization with WaitGroup ===")

	var wg sync.WaitGroup // WaitGroup coordinates goroutine completion

	// Fork-Join Model:
	// - Fork: Create concurrent execution (go keyword)
	// - Join: Wait for concurrent execution to complete (wg.Wait())

	sayHello := func() {
		defer wg.Done() // Signal completion when function exits
		fmt.Println("hello with proper sync")
	}

	wg.Add(1)     // Tell WaitGroup: "I'm starting 1 goroutine"
	go sayHello() // Fork: Create goroutine
	wg.Wait()     // Join: Block until goroutine calls Done()

	fmt.Println("main goroutine continues after join point")
}

// ============================================================================
// 4. GOROUTINES AND CLOSURES - LEXICAL SCOPE
// ============================================================================

func closuresAndScope() {
	fmt.Println("\n=== Goroutines and Closures ===")

	var wg sync.WaitGroup

	// Example 1: Goroutines operate in the same address space
	salutation := "hello"
	wg.Add(1)
	wg.Go(func() {
		defer wg.Done()
		salutation = "welcome" // Modifies the original variable
	})
	fmt.Printf("Before goroutine: %s\n", salutation) // Print: "Hello"
	wg.Wait()
	fmt.Printf("After goroutine: %s\n", salutation) // Prints: "welcome"
}

// ============================================================================
// 5. COMMON PITFALL: LOOP VARIABLE CAPTURE
// ============================================================================

func loopVariablePitfall() {
	fmt.Println("\n=== Loop Variable Pitfall (WRONG) ===")

	var wg sync.WaitGroup

	var salutation string
	// WRONG: All goroutines reference the same loop variable
	for _, s := range []string{"hello", "greetings", "good day"} {
		salutation = s
		wg.Go(func() {
			// By the time this runs, salutation likely = "good day" (last value)
			// time.Sleep(1 * time.Millisecond)
			fmt.Println(salutation)
		})
	}
	wg.Wait()
	// Often prints "good day" three times!
}

func loopVariableFixed() {
	fmt.Println("\n=== Loop Variable Fixed (CORRECT) ===")

	var wg sync.WaitGroup
	var salutation string

	// CORRECT: Pass a copy of the loop variable to each goroutine
	for _, s := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		salutation = s
		go func(s string) { // Parameter creates a copy
			defer wg.Done()
			fmt.Println(s)
		}(salutation) // Pass current value as argument
	}
	wg.Wait()
	// Correctly prints all three values (in random order)
}

// ============================================================================
// 6. GOROUTINES ARE NOT GARBAGE COLLECTED
// ============================================================================

func goroutineLeaks() {
	fmt.Println("\n=== Goroutine Leaks ===")

	// WARNING: This goroutine will never exit and won't be garbage collected
	go func() {
		for {
			// Infinite loop - goroutine leaks until process exits
			time.Sleep(1 * time.Second)
		}
	}()

	// The goroutine above will hang around until the program terminates
	// Always ensure goroutines have a way to exit!
}

// ============================================================================
// 7. MEASURING GOROUTINE SIZE
// ============================================================================

func measureGoroutineSize() {
	fmt.Println("\n=== Measuring Goroutine Memory ===")

	memConsumed := func() uint64 {
		runtime.GC() // Force garbage collection
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys // Total memory from OS
	}

	var c <-chan interface{} // Channel that will never send (blocks forever)
	var wg sync.WaitGroup
	noop := func() {
		wg.Done()
		<-c // Block forever to keep goroutine alive for measurement
	}

	const numGoroutines = 1e4 // 10,000 goroutines
	wg.Add(numGoroutines)

	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}
	wg.Wait()

	after := memConsumed()

	// Calculate average size per goroutine
	avgSize := float64(after-before) / numGoroutines / 1000
	fmt.Printf("Average goroutine size: %.3fkb\n", avgSize)
	fmt.Printf("Created %.3f goroutines\n", numGoroutines)

	// Result: ~2-3kb per goroutine (very lightweight!)
}

// ============================================================================
// 8. M:N SCHEDULER DEMONSTRATION
// ============================================================================
// **Characteristics:**
// - **Lightweight**: 2-8 KB initial stack size (grows dynamically)
// - **Cheap creation**: No system calls (~microseconds)
// - **Runtime scheduled**: Go scheduler (not OS) decides when to run
// - **Cooperative + Preemptive**: Go runtime manages switching
// - **Massive scalability**: Millions of goroutines possible
// - **M:N mapping**: Many goroutines multiplexed onto few OS threads

// **Example cost:**
// - 100,000 goroutines ≈ 200 MB - 800 MB memory

// ## The M:N Model (Go's Approach)

// Go uses an **M:N scheduler**: M goroutines run on N OS threads (where M >> N).
// ```
// Goroutines (hundreds of thousands):
//
//	G1  G2  G3  G4  G5  G6  G7  G8  ...  G100000
//	 \   |   /    \   |   /    \   |
//	  \  |  /      \  |  /      \  |
//	   Machine (OS Thread) Pool:
//	     M1        M2        M3       M4
//	      \         |         /       /
//	       \        |        /       /
//	        Operating System Kernel
func schedulerDemo() {
	fmt.Println("\n=== M:N Scheduler (Goroutines on OS Threads) ===")

	// Go uses an M:N scheduler:
	// M goroutines mapped to N OS threads (M >> N)

	numCPU := runtime.NumCPU()
	numGoroutine := runtime.NumGoroutine()

	fmt.Printf("Number of CPUs: %d\n", numCPU)
	fmt.Printf("Number of OS threads (GOMAXPROCS): %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Current goroutines running: %d\n", numGoroutine)

	// Create many goroutines - they'll be multiplexed onto few OS threads
	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	fmt.Printf("After creating 1000 goroutines: %d running\n", runtime.NumGoroutine())
	wg.Wait()
}

// ============================================================================
// 9. COROUTINES VS GOROUTINES
// ============================================================================

func coroutineExplanation() {
	fmt.Println("\n=== Goroutines as Special Coroutines ===")

	// Goroutines are a special class of coroutines:
	// - Coroutines: Concurrent subroutines that are NON-PREEMPTIVE
	// - They cannot be interrupted arbitrarily
	// - They have suspension/reentry points

	// What makes goroutines special:
	// - Go runtime automatically manages suspension/reentry
	// - Goroutines are suspended when they BLOCK (I/O, channel ops, etc.)
	// - Goroutines resume when they become UNBLOCKED
	// - This makes them "cooperatively scheduled" but runtime-managed

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Will block on channel receive
	go func() {
		defer wg.Done()
		c := make(chan int)
		fmt.Println("Goroutine 1: Waiting on channel (blocked)...")
		<-c // Blocks here - runtime suspends this goroutine
	}()

	// Goroutine 2: Runs while goroutine 1 is blocked
	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 2: Running while 1 is blocked")
		time.Sleep(10 * time.Millisecond)
	}()

	time.Sleep(50 * time.Millisecond)
	// Note: Goroutine 1 will leak (blocked forever) - for demonstration only
}

// ============================================================================
// 10. SCALABILITY: CREATING MANY GOROUTINES
// ============================================================================

func scalabilityDemo() {
	fmt.Println("\n=== Scalability: Creating 100,000 Goroutines ===")

	var wg sync.WaitGroup
	const count = 100000

	start := time.Now()

	for i := range count {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine does minimal work
			_ = id * 2
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Created and completed %d goroutines in %v\n", count, elapsed)
	fmt.Println("Goroutines are extremely lightweight and scalable!")
}

// ============================================================================
// 11. CONTEXT SWITCHING PERFORMANCE
// ============================================================================

func contextSwitchingDemo() {
	fmt.Println("\n=== Context Switching Performance ===")

	// Context switching between goroutines is much faster than OS threads
	// - OS thread context switch: ~1-2 microseconds
	// - Goroutine context switch: ~0.2 microseconds (10x faster!)

	var wg sync.WaitGroup
	c := make(chan struct{})

	sender := func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			c <- struct{}{} // Send message
		}
		close(c)
	}

	receiver := func() {
		defer wg.Done()
		for range c { // Receive messages
			// Context switches happen here as goroutines alternate
		}
	}

	wg.Add(2)
	start := time.Now()

	go sender()
	go receiver()

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("100,000 context switches in %v\n", elapsed)
	fmt.Println("Software-level scheduling is much faster than OS-level!")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func goRoutine() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           COMPLETE GOROUTINES GUIDE WITH EXAMPLES          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run each demonstration
	basicGoroutine()
	anonymousGoroutines()
	properSynchronization()
	closuresAndScope()
	loopVariablePitfall()
	loopVariableFixed()
	goroutineLeaks()
	measureGoroutineSize()
	schedulerDemo()
	coroutineExplanation()
	scalabilityDemo()
	contextSwitchingDemo()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY TAKEAWAYS                             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║ 1. Always use sync.WaitGroup for proper synchronization      ║")
	fmt.Println("║ 2. Pass loop variables as parameters to goroutines           ║")
	fmt.Println("║ 3. Goroutines are ~2-3KB each (extremely lightweight)        ║")
	fmt.Println("║ 4. Can create millions of goroutines easily                  ║")
	fmt.Println("║ 5. Context switching is 10x faster than OS threads           ║")
	fmt.Println("║ 6. Goroutines are NOT garbage collected - avoid leaks        ║")
	fmt.Println("║ 7. Use 'go' keyword to fork, WaitGroup.Wait() to join        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}
