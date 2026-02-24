package syncpackage

import (
	"bytes"
	"fmt"

	// "io"
	"sync"
	"time"
)

// ============================================================================
// sync.Pool - COMPLETE GUIDE
// ============================================================================
// sync.Pool = Concurrent-safe object pool pattern
// - Reuse objects instead of creating new ones (reduces GC pressure)
// - Thread-safe for multiple goroutines
// - Objects can be evicted at any time (GC can clear the pool)
// - Best for frequently allocated/deallocated objects
// ============================================================================

// ============================================================================
// 1. BASIC sync.Pool USAGE
// ============================================================================

func basicPool() {
	fmt.Println("\n=== Basic sync.Pool Usage ===")

	// Create a pool with a New function
	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("  Creating new instance")
			return struct{}{}
		},
	}

	// First Get - no instances in pool, calls New
	fmt.Println("First Get():")
	myPool.Get()

	// Second Get - still no instances, calls New again
	fmt.Println("\nSecond Get():")
	instance := myPool.Get()

	// Put instance back in pool
	fmt.Println("\nPut() instance back:")
	myPool.Put(instance)

	// Third Get - reuses instance from pool, no New call!
	fmt.Println("\nThird Get() (reuses from pool):")
	myPool.Get()

	fmt.Println("\nOnly 2 'Creating new instance' messages (not 3)!")
}

// ============================================================================
// 2. WHY USE sync.Pool? MEMORY OPTIMIZATION
// ============================================================================

func memoryOptimization() {
	fmt.Println("\n=== Memory Optimization with sync.Pool ===")

	var numCalcsCreated int

	// Pool that creates 1KB byte slices
	calcPool := &sync.Pool{
		New: func() interface{} {
			numCalcsCreated++
			mem := make([]byte, 1024) // 1KB
			return &mem               // IMPORTANT: Store pointer, not value
		},
	}

	// Seed the pool with 4 instances (4KB total)
	fmt.Println("Seeding pool with 4 instances (4KB)...")
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	// Simulate 1 million workers
	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	fmt.Printf("Starting %d workers...\n", numWorkers)
	start := time.Now()

	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			// Get from pool (type assertion)
			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem) // ALWAYS Put back!

			// Simulate quick operation with memory
			_ = (*mem)[0]
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\nCompleted in %v\n", elapsed)
	fmt.Printf("%d objects created (not %d!)\n", numCalcsCreated, numWorkers)
	fmt.Printf("Without pool: ~1GB memory\n")
	fmt.Printf("With pool: ~%dKB memory (reused objects)\n", numCalcsCreated)
}

// ============================================================================
// 3. WITHOUT POOL VS WITH POOL
// ============================================================================

func withoutPool() {
	fmt.Println("\n=== WITHOUT sync.Pool (Allocates Every Time) ===")

	var created int

	var wg sync.WaitGroup
	const workers = 10000
	wg.Add(workers)

	start := time.Now()

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			// Create new buffer every time
			buffer := make([]byte, 1024)
			created++

			// Use buffer
			_ = buffer[0]
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Created %d buffers in %v\n", workers, elapsed)
	fmt.Println("→ More GC pressure, more allocations")
}

func withPool() {
	fmt.Println("\n=== WITH sync.Pool (Reuses Objects) ===")

	var created int

	pool := &sync.Pool{
		New: func() interface{} {
			created++
			buffer := make([]byte, 1024)
			return &buffer
		},
	}

	// Warm the pool
	for i := 0; i < 10; i++ {
		pool.Put(pool.New())
	}

	var wg sync.WaitGroup
	const workers = 10000
	wg.Add(workers)

	start := time.Now()

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			// Get from pool (reuses existing)
			buffer := pool.Get().(*[]byte)
			defer pool.Put(buffer)

			// Use buffer
			_ = (*buffer)[0]
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Created only %d buffers (reused!) in %v\n", created, elapsed)
	fmt.Println("→ Less GC pressure, fewer allocations")
}

// ============================================================================
// 4. CACHE WARMING FOR PERFORMANCE
// ============================================================================

func cacheWarming() {
	fmt.Println("\n=== Cache Warming for High-Performance ===")

	// Simulate expensive object creation
	expensiveNew := func() interface{} {
		time.Sleep(10 * time.Millisecond) // Expensive!
		return &struct{ data string }{data: "expensive"}
	}

	fmt.Println("\nWithout pre-warming:")
	pool1 := &sync.Pool{New: expensiveNew}

	start := time.Now()
	for i := 0; i < 5; i++ {
		obj := pool1.Get()
		pool1.Put(obj)
	}
	fmt.Printf("Time: %v (each Get() calls New)\n", time.Since(start))

	fmt.Println("\nWith pre-warming:")
	pool2 := &sync.Pool{New: expensiveNew}

	// Pre-warm the pool
	for i := 0; i < 5; i++ {
		pool2.Put(expensiveNew())
	}

	start = time.Now()
	for i := 0; i < 5; i++ {
		obj := pool2.Get()
		pool2.Put(obj)
	}
	fmt.Printf("Time: %v (gets from warm cache)\n", time.Since(start))

	fmt.Println("\nPre-warming is crucial for performance-critical code!")
}

// ============================================================================
// 5. REAL-WORLD EXAMPLE: BUFFER POOL
// ============================================================================

// BufferPool demonstrates a common real-world use case
type BufferPool struct {
	pool *sync.Pool
}

func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				// Create 4KB buffers
				return new(bytes.Buffer)
			},
		},
	}
}

func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	// IMPORTANT: Reset buffer before putting back
	buf.Reset()
	bp.pool.Put(buf)
}

func bufferPoolExample() {
	fmt.Println("\n=== Real-World: Buffer Pool ===")

	pool := NewBufferPool()

	var wg sync.WaitGroup

	// Simulate multiple goroutines using buffers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get buffer from pool
			buf := pool.Get()
			defer pool.Put(buf) // Return to pool when done

			// Use buffer
			fmt.Fprintf(buf, "Worker %d: Hello, World!", id)
			fmt.Printf("%s\n", buf.String())
		}(i)
	}

	wg.Wait()
	fmt.Println("Buffers reused efficiently across goroutines!")
}

// ============================================================================
// 6. WHEN TO USE sync.Pool
// ============================================================================

func whenToUse() {
	fmt.Println("\n=== When to Use sync.Pool ===")

	fmt.Println("\n✅ USE sync.Pool when:")
	fmt.Println("  • Objects are expensive to create")
	fmt.Println("  • Objects are frequently allocated and deallocated")
	fmt.Println("  • Objects are roughly uniform (same type/size)")
	fmt.Println("  • You want to reduce GC pressure")
	fmt.Println("  • High-throughput scenarios (servers, APIs)")

	fmt.Println("\n❌ DON'T use sync.Pool when:")
	fmt.Println("  • Objects have variable sizes (e.g., random-length slices)")
	fmt.Println("  • Objects are rarely reused")
	fmt.Println("  • Objects are cheap to create")
	fmt.Println("  • Objects require complex initialization")
	fmt.Println("  • You need guaranteed object availability")

	fmt.Println("\n📝 Common Use Cases:")
	fmt.Println("  • Buffer pools (bytes.Buffer, strings.Builder)")
	fmt.Println("  • HTTP request/response objects")
	fmt.Println("  • JSON encoders/decoders")
	fmt.Println("  • Database connection pre-allocation")
	fmt.Println("  • Temporary data structures in hot paths")
}

// ============================================================================
// 7. BEST PRACTICES AND WARNINGS
// ============================================================================

func poolBestPractices() {
	fmt.Println("\n=== sync.Pool Best Practices ===")

	fmt.Println("\n1. New function must be thread-safe:")
	goodPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer) // ✓ Thread-safe
		},
	}
	_ = goodPool

	fmt.Println("\n2. ALWAYS Put objects back:")
	fmt.Println("   buffer := pool.Get().(*bytes.Buffer)")
	fmt.Println("   defer pool.Put(buffer) // ✓ Use defer")

	fmt.Println("\n3. Reset object state before Put:")
	pool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	buf := pool.Get().(*bytes.Buffer)
	buf.WriteString("data")
	buf.Reset() // ✓ Clear state
	pool.Put(buf)

	fmt.Println("\n4. No assumptions about retrieved objects:")
	buf2 := pool.Get().(*bytes.Buffer)
	// Don't assume buf2 is empty! It might have stale data
	buf2.Reset() // ✓ Always reset/initialize
	pool.Put(buf2)

	fmt.Println("\n5. Store pointers, not values:")
	fmt.Println("   ✓ return &buffer  (pointer)")
	fmt.Println("   ✗ return buffer   (value gets copied)")

	fmt.Println("\n⚠️  Objects can be evicted by GC at any time!")
	fmt.Println("    Don't rely on Pool for persistent storage")
}

// ============================================================================
// 8. COMMON PITFALLS
// ============================================================================

func commonPitfalls() {
	fmt.Println("\n=== Common Pitfalls ===")

	fmt.Println("\n❌ Pitfall 1: Not resetting state")
	pool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	buf := pool.Get().(*bytes.Buffer)
	buf.WriteString("secret data")
	pool.Put(buf) // BUG: Didn't reset!

	buf2 := pool.Get().(*bytes.Buffer)
	fmt.Printf("  Got buffer with stale data: %q\n", buf2.String())

	fmt.Println("\n❌ Pitfall 2: Wrong type assertion")
	fmt.Println("  obj := pool.Get().(WrongType) // PANIC!")

	fmt.Println("\n❌ Pitfall 3: Forgetting to Put back")
	fmt.Println("  buffer := pool.Get()")
	fmt.Println("  // forgot pool.Put(buffer) - Pool becomes useless!")

	fmt.Println("\n❌ Pitfall 4: Variable-sized objects")
	fmt.Println("  If you need slices of length 10, 100, 1000...")
	fmt.Println("  Pool won't help - you'll rarely get the right size")

	fmt.Println("\n❌ Pitfall 5: Storing values instead of pointers")
	badPool := &sync.Pool{
		New: func() interface{} {
			buffer := make([]byte, 1024)
			return buffer // BUG: Returns value, not pointer
		},
	}
	_ = badPool
	fmt.Println("  Values get copied, defeating the purpose!")
}

// ============================================================================
// 9. PERFORMANCE BENCHMARK SIMULATION
// ============================================================================

func performanceBenchmark() {
	fmt.Println("\n=== Performance Comparison ===")

	// Simulate expensive object creation
	createExpensive := func() *bytes.Buffer {
		time.Sleep(1 * time.Microsecond) // Simulate work
		return new(bytes.Buffer)
	}

	const operations = 10000

	// Without Pool
	fmt.Println("\nWithout Pool:")
	start := time.Now()
	for i := 0; i < operations; i++ {
		buf := createExpensive()
		_ = buf
		// Object becomes garbage
	}
	withoutPool := time.Since(start)
	fmt.Printf("  Time: %v\n", withoutPool)

	// With Pool
	fmt.Println("\nWith Pool:")
	pool := &sync.Pool{
		New: func() interface{} {
			time.Sleep(1 * time.Microsecond)
			return new(bytes.Buffer)
		},
	}

	// Warm pool
	for i := 0; i < 10; i++ {
		pool.Put(pool.New())
	}

	start = time.Now()
	for i := 0; i < operations; i++ {
		buf := pool.Get().(*bytes.Buffer)
		pool.Put(buf)
		// Object reused
	}
	withPool := time.Since(start)
	fmt.Printf("  Time: %v\n", withPool)

	if withoutPool > withPool {
		speedup := float64(withoutPool) / float64(withPool)
		fmt.Printf("\nPool is %.2fx faster!\n", speedup)
	}
}

// ============================================================================
// 10. REAL-WORLD: HTTP SERVER WITH POOL
// ============================================================================

type ResponseWriter struct {
	buffer *bytes.Buffer
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	return rw.buffer.Write(data)
}

func (rw *ResponseWriter) String() string {
	return rw.buffer.String()
}

func httpServerExample() {
	fmt.Println("\n=== Real-World: HTTP Response Writer Pool ===")

	// Pool of response writers
	writerPool := &sync.Pool{
		New: func() interface{} {
			return &ResponseWriter{
				buffer: new(bytes.Buffer),
			}
		},
	}

	// Simulate handling requests
	handleRequest := func(id int) {
		// Get writer from pool
		writer := writerPool.Get().(*ResponseWriter)
		defer func() {
			// Reset and return to pool
			writer.buffer.Reset()
			writerPool.Put(writer)
		}()

		// Use writer
		fmt.Fprintf(writer.buffer, "Response to request %d", id)
		fmt.Printf("  %s\n", writer.String())
	}

	var wg sync.WaitGroup

	// Simulate 10 concurrent requests
	fmt.Println("Handling 10 requests with pooled writers:")
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handleRequest(id)
		}(i)
	}

	wg.Wait()
	fmt.Println("All requests handled with object reuse!")
}

// ============================================================================
// 11. ADVANCED: TYPED POOL WRAPPER
// ============================================================================

// TypedPool wraps sync.Pool with type safety
type TypedPool[T any] struct {
	pool *sync.Pool
}

func NewTypedPool[T any](newFunc func() *T) *TypedPool[T] {
	return &TypedPool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

func (p *TypedPool[T]) Get() *T {
	return p.pool.Get().(*T)
}

func (p *TypedPool[T]) Put(item *T) {
	p.pool.Put(item)
}

func typedPoolExample() {
	fmt.Println("\n=== Type-Safe Pool Wrapper (Go 1.18+) ===")

	// Create type-safe buffer pool
	bufferPool := NewTypedPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	// No type assertions needed!
	buffer := bufferPool.Get() // Returns *bytes.Buffer directly
	buffer.WriteString("Type safe!")
	fmt.Printf("Buffer: %s\n", buffer.String())

	buffer.Reset()
	bufferPool.Put(buffer)

	fmt.Println("Type safety eliminates runtime panics!")
}

// ============================================================================
// 12. POOL VS OTHER PATTERNS
// ============================================================================

func poolVsOthers() {
	fmt.Println("\n=== sync.Pool vs Other Patterns ===")

	fmt.Println("\nsync.Pool vs Manual Pool:")
	fmt.Println("  Manual: var objects = make(chan *Object, 100)")
	fmt.Println("  ✗ Fixed size, can block")
	fmt.Println("  sync.Pool:")
	fmt.Println("  ✓ Dynamic size, never blocks")
	fmt.Println("  ✓ GC can reclaim unused objects")

	fmt.Println("\nsync.Pool vs Always Allocating:")
	fmt.Println("  Always allocate: obj := new(Object)")
	fmt.Println("  ✗ GC pressure, slower")
	fmt.Println("  sync.Pool:")
	fmt.Println("  ✓ Reuses objects, less GC")
	fmt.Println("  ✓ Much faster for hot paths")

	fmt.Println("\nsync.Pool vs Global Variables:")
	fmt.Println("  Global: var globalBuffer bytes.Buffer")
	fmt.Println("  ✗ Needs locking, not concurrent-safe")
	fmt.Println("  sync.Pool:")
	fmt.Println("  ✓ Concurrent-safe")
	fmt.Println("  ✓ Multiple objects available")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func PoolDemo() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              sync.Pool COMPLETE GUIDE                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run all demonstrations
	basicPool()
	memoryOptimization()
	withoutPool()
	withPool()
	cacheWarming()
	bufferPoolExample()
	whenToUse()
	poolBestPractices()
	commonPitfalls()
	performanceBenchmark()
	httpServerExample()
	typedPoolExample()
	poolVsOthers()

	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY TAKEAWAYS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║ sync.Pool = Concurrent-safe object pool                   ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Two Methods:                                               ║")
	fmt.Println("║   • Get() - Retrieve object (calls New if empty)          ║")
	fmt.Println("║   • Put() - Return object to pool for reuse               ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Benefits:                                                  ║")
	fmt.Println("║   • Reduces memory allocations                            ║")
	fmt.Println("║   • Reduces GC pressure                                   ║")
	fmt.Println("║   • Improves performance (can be 100-1000x faster)        ║")
	fmt.Println("║   • Thread-safe                                           ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ GOLDEN RULES:                                              ║")
	fmt.Println("║   1. New function must be thread-safe                     ║")
	fmt.Println("║   2. ALWAYS Put() objects back (use defer)                ║")
	fmt.Println("║   3. Reset object state before Put()                      ║")
	fmt.Println("║   4. Make NO assumptions about Get() results              ║")
	fmt.Println("║   5. Store POINTERS, not values                           ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Use When:                                                  ║")
	fmt.Println("║   ✅ Objects expensive to create                           ║")
	fmt.Println("║   ✅ Frequent allocation/deallocation                      ║")
	fmt.Println("║   ✅ Uniform object sizes                                  ║")
	fmt.Println("║   ✅ High-throughput scenarios                             ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Don't Use When:                                            ║")
	fmt.Println("║   ❌ Variable-sized objects                                ║")
	fmt.Println("║   ❌ Rare reuse                                            ║")
	fmt.Println("║   ❌ Cheap to create                                       ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ ⚠️  IMPORTANT: GC can evict objects anytime!               ║")
	fmt.Println("║     Don't rely on Pool for persistent storage             ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Pattern:                                                   ║")
	fmt.Println("║   pool := &sync.Pool{                                     ║")
	fmt.Println("║       New: func() interface{} {                           ║")
	fmt.Println("║           return new(ExpensiveObject)                     ║")
	fmt.Println("║       },                                                   ║")
	fmt.Println("║   }                                                        ║")
	fmt.Println("║   obj := pool.Get().(*ExpensiveObject)                    ║")
	fmt.Println("║   defer pool.Put(obj)                                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}
