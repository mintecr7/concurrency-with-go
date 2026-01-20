package syncpackage

import (
	"fmt"
	"math"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

// ============================================================================
// MUTEX AND RWMUTEX - COMPLETE GUIDE
// ============================================================================
// A Mutex is a mutual exclusion lock.
// The zero value for a Mutex is an unlocked mutex.
//
// A Mutex must not be copied after first use.
//
// In the terminology of [the Go memory model],
// the n'th call to [Mutex.Unlock] “synchronizes before” the m'th call to [Mutex.Lock]
// for any n < m.
// A successful call to [Mutex.TryLock] is equivalent to a call to Lock.
// A failed call to TryLock does not establish any “synchronizes before”
// relation at all.
//
// Mutex = "Mutual Exclusion"
// - Guards critical sections (exclusive access to shared resources)
// - Memory access synchronization (low-level primitive)
// - "Share memory by creating conventions" (vs channels: "share by communicating")
// ============================================================================

// ============================================================================
// 1. BASIC MUTEX USAGE
// ============================================================================

func basicMutex() {
	fmt.Println("\n=== Basic Mutex Usage ===")

	var count int
	var lock sync.Mutex // Mutex guards the critical section

	increment := func() {
		lock.Lock()         // Request exclusive access
		defer lock.Unlock() // ALWAYS use defer to ensure unlock (even on panic)
		count++
		fmt.Printf("Incrementing: %d\n", count)
	}

	decrement := func() {
		lock.Lock()         // Request exclusive access
		defer lock.Unlock() // Release lock when done
		count--
		fmt.Printf("Decrementing: %d\n", count)
	}

	var arithmetic sync.WaitGroup

	// Launch 5 increment goroutines
	for range 5 {
		arithmetic.Go(func() {
			increment()
		})
	}

	// Launch 5 decrement goroutines
	for range 5 {
		arithmetic.Go(func() {
			decrement()
		})
	}

	arithmetic.Wait()
	fmt.Printf("Final count: %d\n", count)
	fmt.Println("Arithmetic complete.")
}

// ============================================================================
// 2. WHY MUTEX IS NEEDED - RACE CONDITION DEMO
// ============================================================================

func withoutMutex() {
	fmt.Println("\n=== WITHOUT Mutex (Race Condition) ===")

	var count int
	// NO MUTEX - This will cause a race condition!

	var wg sync.WaitGroup

	for range 1000 {
		wg.Go(func() {
			count++ // UNSAFE: Multiple goroutines writing simultaneously
		})
	}

	wg.Wait()
	fmt.Printf("Expected: 1000, Got: %d (likely wrong due to race)\n", count)
	fmt.Println("Run with: go run -race <file> to detect race conditions")
}

func withMutex() {
	fmt.Println("\n=== WITH Mutex (Safe) ===")

	var count int
	var mu sync.Mutex // Mutex protects count

	var wg sync.WaitGroup

	for range 1000 {
		wg.Go(func() {
			mu.Lock()   // Acquire lock
			count++     // SAFE: Only one goroutine at a time
			mu.Unlock() // Release lock
		})
	}

	wg.Wait()
	fmt.Printf("Expected: 1000, Got: %d (correct!)\n", count)
}

// ============================================================================
// 3. CRITICAL SECTIONS EXPLAINED
// ============================================================================

func criticalSections() {
	fmt.Println("\n=== Critical Sections ===")

	// Critical section = code that requires exclusive access to shared resource
	// They're called "critical" because they're bottlenecks

	var balance int = 1000
	var mu sync.Mutex

	withdraw := func(amount int) bool {
		mu.Lock()
		defer mu.Unlock()

		// === CRITICAL SECTION START ===
		// Only one goroutine can execute this at a time
		if balance >= amount {
			time.Sleep(1 * time.Millisecond) // Simulate processing
			balance -= amount
			fmt.Printf("Withdrew $%d, balance: $%d\n", amount, balance)
			return true
		}
		// === CRITICAL SECTION END ===

		fmt.Printf("Insufficient funds for $%d\n", amount)
		return false
	}

	var wg sync.WaitGroup

	// Multiple goroutines trying to withdraw
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			withdraw(200)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Final balance: $%d\n", balance)
}

// ============================================================================
// 4. MUTEX BEST PRACTICES
// ============================================================================

func mutexBestPractices() {
	fmt.Println("\n=== Mutex Best Practices ===")

	fmt.Println("\n✓ DO:")
	fmt.Println("1. ALWAYS use 'defer mu.Unlock()' immediately after Lock()")
	fmt.Println("2. Keep critical sections as SHORT as possible")
	fmt.Println("3. Lock at the smallest scope necessary")
	fmt.Println("4. Document what the mutex protects")

	fmt.Println("\n✗ DON'T:")
	fmt.Println("1. Forget to unlock (causes deadlock)")
	fmt.Println("2. Lock for long operations (reduces concurrency)")
	fmt.Println("3. Copy a Mutex (pass by pointer only)")
	fmt.Println("4. Lock recursively (same goroutine locks twice = deadlock)")
}

// ============================================================================
// 5. REDUCING CRITICAL SECTION SIZE
// ============================================================================

func criticalSectionOptimization() {
	fmt.Println("\n=== Optimizing Critical Section Size ===")

	var data []int
	var mu sync.Mutex

	// BAD: Large critical section
	badAppend := func(value int) {
		mu.Lock()
		defer mu.Unlock()

		// Expensive operation inside lock
		time.Sleep(1 * time.Millisecond) // Simulate work
		result := value * 2
		data = append(data, result) // Only this needs protection!
	}

	// GOOD: Minimal critical section
	goodAppend := func(value int) {
		// Do expensive work OUTSIDE the lock
		time.Sleep(1 * time.Millisecond) // Simulate work
		result := value * 2

		// Lock only when accessing shared data
		mu.Lock()
		data = append(data, result)
		mu.Unlock()
	}

	fmt.Println("BAD: Large critical section = less concurrency")
	fmt.Println("GOOD: Small critical section = more concurrency")
	_ = badAppend
	_ = goodAppend
}

// ============================================================================
// 6. RWMUTEX (Read-Write Mutex)
// ============================================================================

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// A RWMutex must not be copied after first use.
//
// If any goroutine calls [RWMutex.Lock] while the lock is already held by
// one or more readers, concurrent calls to [RWMutex.RLock] will block until
// the writer has acquired (and released) the lock, to ensure that
// the lock eventually becomes available to the writer.
// Note that this prohibits recursive read-locking.
// A [RWMutex.RLock] cannot be upgraded into a [RWMutex.Lock],
// nor can a [RWMutex.Lock] be downgraded into a [RWMutex.RLock].

func basicRWMutex() {
	fmt.Println("\n=== RWMutex (Read-Write Mutex) ===")

	// RWMutex allows multiple readers OR one writer
	// - Multiple readers can hold RLock() simultaneously
	// - Only one writer can hold Lock() (exclusive)
	// - Writer blocks all readers and other writers

	var data = map[string]int{"counter": 0}
	var mu sync.RWMutex

	// Reader: Uses RLock() for read-only access
	read := func(key string) int {
		mu.RLock() // Multiple readers can hold this simultaneously
		defer mu.RUnlock()

		value := data[key]
		time.Sleep(1 * time.Millisecond)
		return value
	}

	// Writer: Uses Lock() for exclusive access
	write := func(key string, value int) {
		mu.Lock() // Exclusive access (blocks all readers and writers)
		defer mu.Unlock()

		data[key] = value
		time.Sleep(1 * time.Millisecond)
	}

	var wg sync.WaitGroup

	// Start 1 writer
	wg.Go(func() {
		for i := range 3 {
			write("counter", i)
			fmt.Printf("Writer: set counter to %d\n", i)
		}
	})

	// Start 5 readers (can read concurrently!)
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range 3 {
				val := read("counter")
				fmt.Printf("Reader %d: read %d\n", id, val)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Multiple readers can read simultaneously with RWMutex!")
}

// ============================================================================
// 7. MUTEX VS RWMUTEX PERFORMANCE COMPARISON
// ============================================================================

func performanceComparison() {
	fmt.Println("\n=== Mutex vs RWMutex Performance ===")

	// Producer: writes occasionally
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 5; i > 0; i-- {
			l.Lock()
			defer l.Unlock()                 // remove defer for demo
			time.Sleep(1 * time.Millisecond) // Less active than observers
		}
	}

	// Observer: reads frequently
	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	// Test function
	test := func(count int, mutex, rwMutex sync.Locker) time.Duration {
		var wg sync.WaitGroup
		wg.Add(count + 1)
		beginTestTime := time.Now()
		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rwMutex)
		}
		wg.Wait()
		return time.Since(beginTestTime)
	}

	// Setup tabular output
	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")

	// Test with increasing number of readers
	for i := range 10 {
		count := int(math.Pow(2, float64(i)))
		rwTime := test(count, &m, m.RLocker()) // Uses RLock for readers
		mutexTime := test(count, &m, &m)       // Uses Lock for readers

		fmt.Fprintf(tw, "%d\t%v\t%v\n", count, rwTime, mutexTime)
	}

	fmt.Println("\nRWMutex becomes faster when many readers (low write ratio)")
}

// ============================================================================
// 8. WHEN TO USE MUTEX VS RWMUTEX
// ============================================================================

func whenToUseWhich() {
	fmt.Println("\n=== When to Use Mutex vs RWMutex ===")

	fmt.Println("\nUse MUTEX when:")
	fmt.Println("  • Reads and writes are roughly equal")
	fmt.Println("  • Critical sections are very short")
	fmt.Println("  • Simplicity is preferred")

	fmt.Println("\nUse RWMUTEX when:")
	fmt.Println("  • Many readers, few writers")
	fmt.Println("  • Read operations are expensive/slow")
	fmt.Println("  • You need concurrent reads")

	fmt.Println("\nRWMutex overhead:")
	fmt.Println("  • More complex internally")
	fmt.Println("  • Only worth it with enough readers (usually 8+)")
}

// ============================================================================
// 9. REAL-WORLD EXAMPLE: CACHE WITH RWMUTEX
// ============================================================================

// Cache demonstrates practical RWMutex usage
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

// Get uses RLock (multiple goroutines can read concurrently)
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock() // Read lock
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	return value, exists
}

// Set uses Lock (exclusive access for writing)
func (c *Cache) Set(key, value string) {
	c.mu.Lock() // Write lock (exclusive)
	defer c.mu.Unlock()

	c.data[key] = value
}

func cacheExample() {
	fmt.Println("\n=== Real-World: Cache with RWMutex ===")

	cache := NewCache()
	var wg sync.WaitGroup

	// 1 writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key%d", i)
			cache.Set(key, fmt.Sprintf("value%d", i))
			fmt.Printf("Writer: Set %s\n", key)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// 10 reader goroutines (can all read simultaneously!)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				key := fmt.Sprintf("key%d", j)
				if value, exists := cache.Get(key); exists {
					fmt.Printf("Reader %d: Got %s=%s\n", id, key, value)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Cache example complete - many readers, few writers!")
}

// ============================================================================
// 10. DEADLOCK EXAMPLES (COMMON MISTAKES)
// ============================================================================

func deadlockExamples() {
	fmt.Println("\n=== Common Deadlock Scenarios ===")

	// Scenario 1: Forgetting to unlock
	fmt.Println("\n1. Forgetting to unlock (COMMENTED - would hang):")
	fmt.Println("   mu.Lock()")
	fmt.Println("   // forgot mu.Unlock() - DEADLOCK!")

	// Scenario 2: Double locking
	fmt.Println("\n2. Recursive locking (COMMENTED - would hang):")
	fmt.Println("   mu.Lock()")
	fmt.Println("   mu.Lock() // Same goroutine - DEADLOCK!")

	// Scenario 3: Lock ordering
	fmt.Println("\n3. Lock ordering problem:")
	var mu1, mu2 sync.Mutex

	// Goroutine 1: locks mu1 then mu2
	go func() {
		mu1.Lock()
		time.Sleep(1 * time.Millisecond)
		mu2.Lock()         // Waits for mu2
		defer mu2.Unlock() // remove defer for demo
		mu1.Unlock()
	}()

	// Goroutine 2: locks mu2 then mu1
	go func() {
		mu2.Lock()
		time.Sleep(1 * time.Millisecond)
		mu1.Lock()         // Waits for mu1 - DEADLOCK!
		defer mu1.Unlock() // remove defer for demo
		mu2.Unlock()
	}()

	fmt.Println("   Solution: Always acquire locks in the same order")

	time.Sleep(5 * time.Millisecond)
}

// ============================================================================
// 11. SYNC.LOCKER INTERFACE
// ============================================================================

func lockerInterface() {
	fmt.Println("\n=== sync.Locker Interface ===")

	// Both Mutex and RWMutex implement sync.Locker
	// sync.Locker interface:
	//   type Locker interface {
	//       Lock()
	//       Unlock()
	//   }

	processWithLock := func(l sync.Locker, name string) {
		l.Lock()
		defer l.Unlock()
		fmt.Printf("Processing with %s\n", name)
	}

	var mu sync.Mutex
	var rwMu sync.RWMutex

	processWithLock(&mu, "Mutex")
	processWithLock(&rwMu, "RWMutex")                    // RWMutex.Lock() is exclusive
	processWithLock(rwMu.RLocker(), "RWMutex.RLocker()") // Returns reader lock

	fmt.Println("Both types satisfy sync.Locker interface!")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func MutexAndRWMutex() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           MUTEX & RWMUTEX COMPLETE GUIDE                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run all demonstrations
	// basicMutex()
	// withoutMutex()
	// withMutex()
	// criticalSections()
	// mutexBestPractices()
	// criticalSectionOptimization()
	basicRWMutex()
	// performanceComparison()
	// whenToUseWhich()
	// cacheExample()
	// deadlockExamples()
	// lockerInterface()

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY TAKEAWAYS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║ MUTEX (Mutual Exclusion):                                  ║")
	fmt.Println("║   • Guards critical sections (exclusive access)            ║")
	fmt.Println("║   • Lock() - acquire, Unlock() - release                   ║")
	fmt.Println("║   • ALWAYS use: defer mu.Unlock()                          ║")
	fmt.Println("║   • Keep critical sections SMALL                           ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ RWMUTEX (Read-Write Mutex):                                ║")
	fmt.Println("║   • Multiple readers OR one writer                         ║")
	fmt.Println("║   • RLock()/RUnlock() - read lock (shared)                 ║")
	fmt.Println("║   • Lock()/Unlock() - write lock (exclusive)               ║")
	fmt.Println("║   • Use when: many reads, few writes                       ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Common Pitfalls:                                           ║")
	fmt.Println("║   • Forgetting to unlock → DEADLOCK                        ║")
	fmt.Println("║   • Locking twice in same goroutine → DEADLOCK             ║")
	fmt.Println("║   • Inconsistent lock ordering → DEADLOCK                  ║")
	fmt.Println("║   • Copying mutex (pass by pointer!)                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}
