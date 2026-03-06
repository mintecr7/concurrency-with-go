package syncpackage

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// sync.Once - COMPLETE GUIDE
// ============================================================================
// sync.Once ensures a function is executed exactly once
// - Thread-safe initialization
// - Common pattern in Go (used 70+ times in standard library)
// - Counts calls to Do(), not unique functions
// ============================================================================

// ============================================================================
// 1. BASIC sync.Once USAGE
// ============================================================================

func basicOnce() {
	fmt.Println("\n=== Basic sync.Once Usage ===")

	var count int
	increment := func() {
		count++
		fmt.Println("Increment called!")
	}

	var once sync.Once
	var wg sync.WaitGroup

	// Launch 100 goroutines, all trying to increment
	wg.Add(100)
	for range 100 {
		go func() {
			defer wg.Done()

			once.Do(increment) // Only executes ONCE across all goroutines
		}()
	}

	wg.Wait()
	fmt.Printf("Count is %d (not 100!)\n", count)
	fmt.Println("sync.Once guarantees the function runs exactly once")
}

// ============================================================================
// 2. SYNC.ONCE COUNTS CALLS TO Do(), NOT UNIQUE FUNCTIONS
// ============================================================================

func onceCalls() {
	fmt.Println("\n=== Once Counts Do() Calls, Not Unique Functions ===")

	var count int

	increment := func() {
		count++
		fmt.Println("Increment executed")
	}

	decrement := func() {
		count--
		fmt.Println("Decrement executed")
	}

	var once sync.Once

	once.Do(increment) // First call: executes
	once.Do(decrement) // Second call: IGNORED (even though it's a different function!)

	fmt.Printf("Count: %d (not 0!)\n", count)
	fmt.Println("\nOnce.Do() only executes the FIRST function passed to it")
	fmt.Println("All subsequent Do() calls are ignored, regardless of the function")
}

// ============================================================================
// 3. WHY sync.Once IS NEEDED
// ============================================================================

func whyOnce() {
	fmt.Println("\n=== Why We Need sync.Once ===")

	fmt.Println("\nWithout sync.Once (Race Condition):")
	var initialized bool
	var data string

	initialize := func() {
		if !initialized { // NOT thread-safe!
			time.Sleep(10 * time.Millisecond) // Simulate work
			data = "initialized"
			initialized = true
			fmt.Println("  Initialization happened")
		}
	}

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		wg.Go(func() {
			initialize()
		})
	}
	wg.Wait()
	fmt.Println("  ^ Might initialize multiple times!")

	fmt.Println("\nWith sync.Once (Thread-Safe):")
	var once sync.Once
	data = ""

	initializeOnce := func() {
		time.Sleep(10 * time.Millisecond)
		data = "initialized"
		fmt.Println("  Initialization happened ONCE")
	}

	wg.Add(5)
	for range 5 {
		go func() {
			defer wg.Done()
			once.Do(initializeOnce) // Guaranteed to run exactly once
		}()
	}
	wg.Wait()
	fmt.Printf("Data: %s\n", data)
	fmt.Println("  ✓ Always initializes exactly once")
}

// ============================================================================
// 4. BEST PRACTICE: TIGHT SCOPE
// ============================================================================

func tightScope() {
	fmt.Println("\n=== Best Practice: Wrap sync.Once in Tight Scope ===")

	// BAD: Global once not coupled to specific function
	// var globalOnce sync.Once

	// GOOD: Encapsulate in a type
	type Database struct {
		once sync.Once
		conn string
	}

	Connect := func(db *Database) {
		db.once.Do(func() {
			fmt.Println("  Connecting to database...")
			time.Sleep(50 * time.Millisecond)
			db.conn = "connected"
			fmt.Println("  Connection established")
		})
	}

	db := &Database{}
	var wg sync.WaitGroup

	// Multiple goroutines try to connect
	for i := range 5 {
		// wg.Add(1)
		wg.Go(func() {
			fmt.Printf("Goroutine %d calling Connect()\n", i)
			Connect(db)
		})
		// go func(id int) {
		// 	defer wg.Done()
		// 	fmt.Printf("Goroutine %d calling Connect()\n", id)
		// 	Connect(db) // Only first call actually connects
		// }(i)
	}

	wg.Wait()
	fmt.Println("Database connection established exactly once!")
}

// ============================================================================
// 5. COMMON PATTERN: LAZY INITIALIZATION
// ============================================================================

// Config demonstrates lazy initialization with sync.Once
type Config struct {
	once   sync.Once
	values map[string]string
}

func (c *Config) Load() map[string]string {
	c.once.Do(func() {
		fmt.Println("  Loading configuration (expensive operation)...")
		time.Sleep(100 * time.Millisecond)
		c.values = map[string]string{
			"host": "localhost",
			"port": "8080",
		}
		fmt.Println("  Configuration loaded!")
	})
	return c.values
}

func lazyInitialization() {
	fmt.Println("\n=== Lazy Initialization Pattern ===")

	config := &Config{}
	var wg sync.WaitGroup

	// Multiple goroutines request config
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := config.Load() // First call loads, others wait
			fmt.Printf("Goroutine %d got config: %v\n", id, cfg)
		}(i)
	}

	wg.Wait()
	fmt.Println("Config loaded exactly once, shared by all goroutines!")
}

// ============================================================================
// 6. SINGLETON PATTERN WITH sync.Once
// ============================================================================

// Singleton demonstrates the singleton pattern
type Singleton struct {
	data string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance() *Singleton {
	once.Do(func() {
		fmt.Println("  Creating singleton instance...")
		time.Sleep(50 * time.Millisecond)
		instance = &Singleton{data: "I'm the only one"}
		fmt.Println("  Singleton created!")
	})
	return instance
}

func singletonPattern() {
	fmt.Println("\n=== Singleton Pattern ===")

	var wg sync.WaitGroup

	// Multiple goroutines try to get instance
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s := GetInstance()
			fmt.Printf("Goroutine %d got instance: %p (%s)\n", id, s, s.data)
		}(i)
	}

	wg.Wait()
	fmt.Println("All goroutines got the same instance (same memory address)!")
}

// ============================================================================
// 7. DEADLOCK WITH CIRCULAR DEPENDENCIES
// ============================================================================

func deadlockExample() {
	fmt.Println("\n=== Deadlock: Circular Dependencies (COMMENTED) ===")

	fmt.Println("  WRONG: This would deadlock:")
	fmt.Println("  var onceA, onceB sync.Once")
	fmt.Println("  var initB func()")
	fmt.Println("  initA := func() { onceB.Do(initB) }")
	fmt.Println("  initB = func() { onceA.Do(initA) }")
	fmt.Println("  onceA.Do(initA) // DEADLOCK!")

	fmt.Println("\nWhy? onceA.Do(initA) waits for onceB.Do(initB)")
	fmt.Println("     onceB.Do(initB) waits for onceA.Do(initA)")
	fmt.Println("     → Circular dependency → Deadlock")

	// Actual code commented to prevent hanging:
	/*
		var onceA, onceB sync.Once
		var initB func()
		initA := func() { onceB.Do(initB) }
		initB = func() { onceA.Do(initA) }
		onceA.Do(initA) // This will deadlock
	*/
}

// ============================================================================
// 8. MULTIPLE sync.Once FOR DIFFERENT INITIALIZATIONS
// ============================================================================

type Service_1 struct {
	onceDB    sync.Once
	onceCache sync.Once
	db        string
	cache     string
}

func (s *Service_1) InitDB() {
	s.onceDB.Do(func() {
		fmt.Println("  Initializing database...")
		time.Sleep(50 * time.Millisecond)
		s.db = "db_connected"
	})
}

func (s *Service_1) InitCache() {
	s.onceCache.Do(func() {
		fmt.Println("  Initializing cache...")
		time.Sleep(50 * time.Millisecond)
		s.cache = "cache_connected"
	})
}

func multipleOnce() {
	fmt.Println("\n=== Multiple sync.Once for Different Resources ===")

	service := &Service_1{}
	var wg sync.WaitGroup

	// Some goroutines need DB
	for i := range 3 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service.InitDB()
			fmt.Printf("Goroutine %d: DB ready\n", id)
		}(i)
	}

	// Some goroutines need Cache
	for i := range 3 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service.InitCache()
			fmt.Printf("Goroutine %d: Cache ready\n", id)
		}(i + 3)
	}

	wg.Wait()
	fmt.Println("Each resource initialized exactly once!")
}

// ============================================================================
// 9. ERROR HANDLING WITH sync.Once
// ============================================================================

type ConnectionPool struct {
	once sync.Once
	conn string
	err  error
}

func (cp *ConnectionPool) Connect() error {
	cp.once.Do(func() {
		fmt.Println("  Attempting connection...")
		time.Sleep(50 * time.Millisecond)

		// Simulate connection error
		if time.Now().Unix()%2 == 0 {
			cp.err = fmt.Errorf("connection failed")
			fmt.Println(" x Connection failed")
		} else {
			cp.conn = "connected"
			fmt.Println("  ✓ Connection successful")
		}
	})
	return cp.err
}

func errorHandling() {
	fmt.Println("\n=== Error Handling with sync.Once ===")

	fmt.Println("\n - Important: sync.Once executes exactly once")
	fmt.Println("Even if the function returns an error!")
	fmt.Println("If initialization fails, it stays failed")

	pool := &ConnectionPool{}

	// First call
	err := pool.Connect()
	if err != nil {
		fmt.Printf("First call: %v\n", err)
	} else {
		fmt.Println("First call: Success")
	}

	// Second call - won't retry even if first failed!
	err = pool.Connect()
	if err != nil {
		fmt.Printf("Second call: %v (same error, no retry)\n", err)
	} else {
		fmt.Println("Second call: Success")
	}

	fmt.Println("\n For retry logic, don't use sync.Once!")
}

// ============================================================================
// 10. PERFORMANCE: WHEN TO USE sync.Once
// ============================================================================

func performance() {
	fmt.Println("\n=== Performance Characteristics ===")

	fmt.Println("\nSync.Once is very fast:")
	fmt.Println("  • First call: Executes function (slow if function is slow)")
	fmt.Println("  • Subsequent calls: Just checks atomic flag (extremely fast)")

	var once sync.Once
	expensiveInit := func() {
		time.Sleep(100 * time.Millisecond)
	}

	// First call
	start := time.Now()
	once.Do(expensiveInit)
	fmt.Printf("First call: %v (includes initialization)\n", time.Since(start))

	// Subsequent calls
	start = time.Now()
	for range 10000 {
		once.Do(expensiveInit)
	}
	fmt.Printf("10,000 subsequent calls: %v (just checks flag)\n", time.Since(start))
}

// ============================================================================
// 11. REAL-WORLD USE CASES
// ============================================================================

func realWorldUseCases() {
	fmt.Println("\n=== Real-World Use Cases ===")

	fmt.Println("\n1. Database Connection Pooling")
	fmt.Println("   var once sync.Once")
	fmt.Println("   func GetDB() *sql.DB {")
	fmt.Println("       once.Do(func() { db = sql.Open(...) })")
	fmt.Println("       return db")
	fmt.Println("   }")

	fmt.Println("\n2. Configuration Loading")
	fmt.Println("   Expensive config file parsing happens once")

	fmt.Println("\n3. Logger Initialization")
	fmt.Println("   Single global logger instance")

	fmt.Println("\n4. Template Parsing")
	fmt.Println("   Parse templates once, reuse many times")

	fmt.Println("\n5. Cache Warming")
	fmt.Println("   Pre-populate cache on first access")

	fmt.Println("\nFound 70+ uses in Go's standard library!")
}

// ============================================================================
// 12. COMPARISON: sync.Once vs Other Patterns
// ============================================================================

func comparison() {
	fmt.Println("\n=== sync.Once vs Other Patterns ===")

	fmt.Println("\nManual flag + mutex:")
	fmt.Println("  var initialized bool")
	fmt.Println("  var mu sync.Mutex")
	fmt.Println("  mu.Lock()")
	fmt.Println("  if !initialized { /* init */ ; initialized = true }")
	fmt.Println("  mu.Unlock()")
	fmt.Println("  Verbose, error-prone, locks every time")

	fmt.Println("\nsync.Once:")
	fmt.Println("  var once sync.Once")
	fmt.Println("  once.Do(initialize)")
	fmt.Println("  Clean, efficient, no locks after first call")

	fmt.Println("\nDouble-checked locking:")
	fmt.Println("  if !initialized { mu.Lock(); ... }")
	fmt.Println("  Complex, can be wrong, sync.Once does it internally")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func RunOnceExamples() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              sync.Once COMPLETE GUIDE                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run all demonstrations
	// basicOnce()
	// onceCalls()
	// whyOnce()
	tightScope()
	// lazyInitialization()
	// singletonPattern()
	// deadlockExample()

	// multipleOnce()
	// errorHandling()
	// performance()
	// realWorldUseCases()
	// comparison()

	// fmt.Println()
	// fmt.Println("╔════════════════════════════════════════════════════════════╗")
	// fmt.Println("║                    KEY TAKEAWAYS                           ║")
	// fmt.Println("╠════════════════════════════════════════════════════════════╣")
	// fmt.Println("║ sync.Once guarantees:                                      ║")
	// fmt.Println("║   • Function executes exactly ONCE                         ║")
	// fmt.Println("║   • Thread-safe across all goroutines                      ║")
	// fmt.Println("║   • Subsequent calls do nothing (very fast)                ║")
	// fmt.Println("║                                                            ║")
	// fmt.Println("║ Important Points:                                          ║")
	// fmt.Println("║   • Once.Do() counts calls, not unique functions           ║")
	// fmt.Println("║   • Only the FIRST function passed to Do() executes        ║")
	// fmt.Println("║   • All subsequent Do() calls are ignored                  ║")
	// fmt.Println("║   • No retry mechanism for errors                          ║")
	// fmt.Println("║                                                            ║")
	// fmt.Println("║ Best Practices:                                            ║")
	// fmt.Println("║   • Wrap Once in tight scope (type or small function)      ║")
	// fmt.Println("║   • Couple Once with specific initialization function      ║")
	// fmt.Println("║   • Avoid circular dependencies (causes deadlock)          ║")
	// fmt.Println("║   • Perfect for lazy initialization                        ║")
	// fmt.Println("║                                                            ║")
	// fmt.Println("║ Common Use Cases:                                          ║")
	// fmt.Println("║   • Singleton pattern                                      ║")
	// fmt.Println("║   • Database connection pools                              ║")
	// fmt.Println("║   • Configuration loading                                  ║")
	// fmt.Println("║   • Template parsing                                       ║")
	// fmt.Println("║   • Logger initialization                                  ║")
	// fmt.Println("║                                                            ║")
	// fmt.Println("║ Pattern:                                                   ║")
	// fmt.Println("║   type Resource struct {                                   ║")
	// fmt.Println("║       once sync.Once                                       ║")
	// fmt.Println("║       data string                                          ║")
	// fmt.Println("║   }                                                        ║")
	// fmt.Println("║   func (r *Resource) Init() {                              ║")
	// fmt.Println("║       r.once.Do(func() { /* expensive init */ })           ║")
	// fmt.Println("║   }                                                        ║")
	// fmt.Println("╚════════════════════════════════════════════════════════════╝")
}
