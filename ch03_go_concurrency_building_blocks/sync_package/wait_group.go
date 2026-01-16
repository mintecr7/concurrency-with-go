package syncpackage

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// THE sync PACKAGE - COMPLETE GUIDE
// ============================================================================
// The sync package provides low-level memory access synchronization primitives.
// These are building blocks for higher-level concurrency patterns.
// Use them for small scopes (like within a struct) when appropriate.
// ============================================================================

// A WaitGroup is a counting semaphore typically used to wait
// for a group of goroutines or tasks to finish.
//
// Typically, a main goroutine will start tasks, each in a new
// goroutine, by calling [WaitGroup.Go] and then wait for all tasks to
// complete by calling [WaitGroup.Wait]. For example:
//	var wg sync.WaitGroup
//	wg.Go(task1)
//	wg.Go(task2)
//	wg.Wait()
//
// A WaitGroup may also be used for tracking tasks without using Go to
// start new goroutines by using [WaitGroup.Add] and [WaitGroup.Done].
//
// ============================================================================
// 1. WAITGROUP - BASIC USAGE
// ============================================================================

// basicWaitGroup demonstrates the fundamental WaitGroup pattern
func basicWaitGroup() {
	fmt.Println("\n=== Basic WaitGroup Usage ===")

	var wg sync.WaitGroup

	// WaitGroup is a concurrent-safe counter:
	// - Add(n): Increment counter by n
	// - Done(): Decrement counter by 1
	// - Wait(): Block until counter reaches 0

	// For the sake of clarity I am manually calling Add and Done in this example
	// however to decrease unexpected human error use the method bellow to spin go routines
	// with a wait group
	// wg.Go(func() {
	// 	fmt.Println("1st goroutine sleeping...")
	// 	time.Sleep(100 * time.Millisecond)
	// })

	wg.Add(1) // Increment counter (1 goroutine starting)
	go func() {
		defer wg.Done() // Decrement counter when done
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(100 * time.Millisecond)
	}()

	wg.Add(1) // Increment counter (another goroutine starting)
	go func() {
		defer wg.Done() // Decrement counter when done
		fmt.Println("2nd goroutine sleeping...")
		time.Sleep(200 * time.Millisecond)
	}()

	wg.Wait() // Block until counter = 0 (all goroutines done)
	fmt.Println("All goroutines complete.")
}

// ============================================================================
// 2. WAITGROUP - COMMON PATTERN WITH LOOPS
// ============================================================================

func waitGroupWithLoops() {
	fmt.Println("\n=== WaitGroup with Loops ===")

	// Common pattern: Add all goroutines before the loop
	hello := func(wg *sync.WaitGroup, id int) {
		defer wg.Done() // Always defer Done() at the start
		fmt.Printf("Hello from goroutine #%d!\n", id)
		time.Sleep(10 * time.Millisecond)
	}

	const numGreeters = 5
	var wg sync.WaitGroup

	// Best practice: Add all goroutines at once BEFORE starting them
	wg.Add(numGreeters)

	for i := range numGreeters {
		go hello(&wg, i+1) // Pass WaitGroup by pointer
	}

	wg.Wait()
	fmt.Println("All greeters finished!")
}

// ============================================================================
// 3. RACE CONDITION: WRONG PLACEMENT OF Add()
// ============================================================================

func wrongAddPlacement() {
	fmt.Println("\n=== WRONG: Add() Inside Goroutine (Race Condition) ===")

	var wg sync.WaitGroup

	// WRONG: Add() is inside the goroutine
	for i := range 5 {
		go func(id int) {
			wg.Add(1) // BUG: May not execute before Wait()
			defer wg.Done()
			fmt.Printf("Goroutine %d\n", id)
		}(i)
	}

	// Race condition: Wait() might execute before any Add() calls
	// This could return immediately, terminating before goroutines run!
	// time.Sleep(10 * time.Millisecond) // Hack to "fix" it (still wrong!)
	wg.Wait()
	fmt.Println("^ This pattern is unsafe - don't use it!")
}

func correctAddPlacement() {
	fmt.Println("\n=== CORRECT: Add() Before Starting Goroutine ===")

	var wg sync.WaitGroup

	// CORRECT: Add() is called BEFORE starting the goroutine
	for i := range 5 {
		wg.Add(1) // Guarantee Add() happens before Wait()
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d\n", id)
		}(i)
	}

	wg.Wait() // Safe: All Add() calls completed before this
	fmt.Println("All goroutines completed safely!")
}

// ============================================================================
// 4. WHEN TO USE WAITGROUP
// ============================================================================

func whenToUseWaitGroup() {
	fmt.Println("\n=== When to Use WaitGroup ===")

	// Use WaitGroup when:
	// 1. You don't care about the result of concurrent operations
	// 2. You have other means of collecting results

	// Example 1: Fire-and-forget operations
	var wg sync.WaitGroup

	tasks := []string{"Send email", "Update database", "Log event"}

	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t string) {
			defer wg.Done()
			fmt.Printf("Executing: %s\n", t)
			time.Sleep(10 * time.Millisecond)
			// No return value needed
		}(task)
	}

	wg.Wait()
	fmt.Println("All tasks completed (no results needed)")
}

// ============================================================================
// 5. WAITGROUP WITH SHARED DATA COLLECTION
// ============================================================================

func waitGroupWithResults() {
	fmt.Println("\n=== WaitGroup with Result Collection ===")

	// When you DO need results, use WaitGroup + another sync primitive
	var wg sync.WaitGroup
	var mu sync.Mutex // Protects shared data
	results := make(map[int]int)

	// Calculate squares of numbers 1-10
	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(n int) {
			defer wg.Done()
			square := n * n

			mu.Lock() // Protect shared map
			results[n] = square
			mu.Unlock()
		}(i)
	}

	wg.Wait() // Wait for all calculations

	fmt.Println("Squares calculated:")
	for i := 1; i <= 10; i++ {
		fmt.Printf("%d² = %d  ", i, results[i])
	}
	fmt.Println()
}

// ============================================================================
// 6. WAITGROUP AS STRUCT FIELD
// ============================================================================

// Service demonstrates WaitGroup as a struct field
type Service struct {
	wg     sync.WaitGroup
	active bool
}

func (s *Service) Start() {
	fmt.Println("\n=== WaitGroup in Struct ===")
	s.active = true

	// Start multiple background workers
	for i := 1; i <= 3; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	fmt.Println("Service started with 3 workers")
}

func (s *Service) worker(id int) {
	defer s.wg.Done()
	fmt.Printf("Worker %d started\n", id)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Worker %d finished\n", id)
}

func (s *Service) Stop() {
	fmt.Println("Stopping service...")
	s.active = false
	s.wg.Wait() // Wait for all workers to finish
	fmt.Println("Service stopped gracefully")
}

func structExample() {
	service := &Service{}
	service.Start()
	time.Sleep(50 * time.Millisecond)
	service.Stop()
}

// ============================================================================
// 7. WAITGROUP COUNTER VISUALIZATION
// ============================================================================

func visualizeCounter() {
	fmt.Println("\n=== WaitGroup Counter Visualization ===")

	var wg sync.WaitGroup

	printCounter := func(label string) {
		// Note: There's no public API to read the counter
		// This is just for conceptual understanding
		fmt.Printf("%s\n", label)
	}

	printCounter("Counter: 0")

	wg.Add(3)
	printCounter("After Add(3) -> Counter: 3")

	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 1 executing...")
		time.Sleep(10 * time.Millisecond)
		printCounter("After Done() #1 -> Counter: 2")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 2 executing...")
		time.Sleep(20 * time.Millisecond)
		printCounter("After Done() #2 -> Counter: 1")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 3 executing...")
		time.Sleep(30 * time.Millisecond)
		printCounter("After Done() #3 -> Counter: 0")
	}()

	fmt.Println("Main waiting at Wait()...")
	wg.Wait() // Blocks until counter = 0
	fmt.Println("Wait() returned! Counter reached 0")
}

// ============================================================================
// 8. COMMON MISTAKES WITH WAITGROUP
// ============================================================================

func commonMistakes() {
	fmt.Println("\n=== Common WaitGroup Mistakes ===")

	// Mistake 1: Forgetting to call Done()
	fmt.Println("\nMistake 1: Forgetting Done() - WILL DEADLOCK!")
	fmt.Println("(Commented out to avoid hanging)")
	/*
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			// Missing: defer wg.Done()
			fmt.Println("Working...")
		}()
		wg.Wait() // Deadlock! Counter never reaches 0
	*/

	// Mistake 2: Calling Add() with wrong count
	fmt.Println("\nMistake 2: Wrong Add() count")
	var wg2 sync.WaitGroup
	wg2.Add(5)    // Says 5 goroutines...
	for range 3 { // But only starts 3!
		go func() {
			defer wg2.Done()
		}()
	}
	// wg2.Wait() // Would deadlock! Counter would be 2
	fmt.Println("(Would deadlock if we called Wait())")

	// Mistake 3: Reusing WaitGroup without resetting
	fmt.Println("\nMistake 3: Reusing WaitGroup")
	var wg3 sync.WaitGroup
	wg3.Add(1)
	go func() { defer wg3.Done() }()
	wg3.Wait()

	// Can reuse, but be careful!
	wg3.Add(1)
	go func() { defer wg3.Done() }()
	wg3.Wait()
	fmt.Println("WaitGroup can be reused (but carefully)")
}

// ============================================================================
// 9. WAITGROUP BEST PRACTICES
// ============================================================================

func bestPractices() {
	fmt.Println("\n=== WaitGroup Best Practices ===")

	fmt.Println("\n✓ BEST PRACTICES:")
	fmt.Println("1. Always use 'defer wg.Done()' at the start of goroutine")
	fmt.Println("2. Call Add() BEFORE starting the goroutine")
	fmt.Println("3. Pass WaitGroup by pointer (*sync.WaitGroup)")
	fmt.Println("4. Add all goroutines at once before loop when possible")
	fmt.Println("5. Match Add() count exactly with number of goroutines")
	fmt.Println("6. Use WaitGroup for 'fire and forget' or when collecting results separately")
	fmt.Println("7. For results, prefer channels over WaitGroup + shared memory")

	fmt.Println("\n✗ AVOID:")
	fmt.Println("1. Calling Add() inside the goroutine")
	fmt.Println("2. Forgetting to call Done()")
	fmt.Println("3. Calling Add() with wrong count")
	fmt.Println("4. Copying WaitGroup (always pass by pointer)")
}

// ============================================================================
// 10. REAL-WORLD EXAMPLE: PARALLEL PROCESSING
// ============================================================================

func realWorldExample() {
	fmt.Println("\n=== Real-World: Parallel File Processing ===")

	// Simulate processing multiple files in parallel
	files := []string{
		"document1.txt", "document2.txt", "document3.txt",
		"document4.txt", "document5.txt",
	}

	processFile := func(wg *sync.WaitGroup, filename string) {
		defer wg.Done()

		fmt.Printf("Processing %s...\n", filename)
		time.Sleep(50 * time.Millisecond) // Simulate work
		fmt.Printf("✓ Completed %s\n", filename)
	}

	var wg sync.WaitGroup

	fmt.Println("Starting parallel file processing...")
	start := time.Now()

	wg.Add(len(files))
	for _, file := range files {
		go processFile(&wg, file)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\nAll files processed in %v\n", elapsed)
	fmt.Println("(Sequential would take ~250ms, parallel took ~50ms)")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func WaitGroupDemo() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              sync.WaitGroup COMPLETE GUIDE                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run all demonstrations
	// basicWaitGroup()
	// waitGroupWithLoops()
	// wrongAddPlacement()
	// correctAddPlacement()
	whenToUseWaitGroup()
	// waitGroupWithResults()
	// structExample()
	// visualizeCounter()
	// commonMistakes()
	// bestPractices()
	// realWorldExample()

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY TAKEAWAYS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║ WaitGroup is a concurrent-safe counter:                    ║")
	fmt.Println("║   • Add(n): Increment counter by n                         ║")
	fmt.Println("║   • Done(): Decrement counter by 1                         ║")
	fmt.Println("║   • Wait(): Block until counter = 0                        ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Golden Rule: Call Add() BEFORE starting goroutine          ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Use WaitGroup when:                                        ║")
	fmt.Println("║   • You don't need results from goroutines                 ║")
	fmt.Println("║   • You collect results via other means (shared data)      ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Don't use WaitGroup when:                                  ║")
	fmt.Println("║   • You need results → Use channels instead                ║")
	fmt.Println("║   • Complex coordination → Use channels + select           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}
