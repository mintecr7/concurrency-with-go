package syncpackage

import (
	"fmt"
	"sync"
	"time"
)

// Cond implements a condition variable, a rendezvous point
// for goroutines waiting for or announcing the occurrence
// of an event.
//
// Each Cond has an associated Locker L (often a [*Mutex] or [*RWMutex]),
// which must be held when changing the condition and
// when calling the [Cond.Wait] method.
//
// A Cond must not be copied after first use.
//
// In the terminology of [the Go memory model], Cond arranges that
// a call to [Cond.Broadcast] or [Cond.Signal] “synchronizes before” any Wait call
// that it unblocks.
//
// For many simple use cases, users will be better off using channels than a
// Cond (Broadcast corresponds to closing a channel, and Signal corresponds to
// sending on a channel).
//
// ============================================================================
// sync.Cond - COMPLETE GUIDE
// ============================================================================
// Cond = Condition Variable
// - A rendezvous point for goroutines waiting for or announcing events
// - An "event" is a signal between goroutines (carries no information except "it occurred")
// - Efficiently suspends goroutines until signaled to wake up
// ============================================================================

// ============================================================================
// 1. THE PROBLEM: WAITING FOR A CONDITION (WRONG APPROACHES)
// ============================================================================

func inefficientApproaches() {
	fmt.Println("\n=== Inefficient Approaches (DON'T DO THIS) ===")

	conditionTrue := func() bool {
		// Simulate some condition
		return time.Now().Unix()%2 == 0
	}

	// WRONG APPROACH 1: Busy waiting (infinite loop)
	fmt.Println("\n1. Busy waiting - consumes 100% of one CPU core:")
	fmt.Println("   for conditionTrue() == false {}")
	fmt.Println("   Wastes CPU cycles")

	// WRONG APPROACH 2: Sleep loop
	fmt.Println("\n2. Sleep loop - better but still inefficient:")
	fmt.Println("   for conditionTrue() == false {")
	fmt.Println("       time.Sleep(1*time.Millisecond)")
	fmt.Println("   }")
	fmt.Println("   Too long = poor performance")
	fmt.Println("   Too short = wasted CPU")

	_ = conditionTrue
}

// ============================================================================
// 2. THE SOLUTION: sync.Cond
// ============================================================================

func basicCond() {
	fmt.Println("\n=== Basic Cond Usage ===")

	// Create a condition variable
	c := sync.NewCond(&sync.Mutex{}) // Requires a Locker (Mutex or RWMutex)
	condition := false

	// Goroutine that waits for the condition
	go func() {
		c.L.Lock() // Lock the condition's Locker

		for !condition { // ALWAYS check condition in a loop!
			fmt.Println("Goroutine: Waiting for condition...")
			c.Wait() // Suspends goroutine (unlocks, waits, then re-locks)
		}

		fmt.Println("Goroutine: Condition met! Proceeding...")
		c.L.Unlock() // Unlock when done
	}()

	// Main goroutine signals the condition
	time.Sleep(100 * time.Millisecond)

	c.L.Lock()
	condition = true // Change the condition
	fmt.Println("Main: Condition set to true, signaling...")
	c.Signal() // Wake up one waiting goroutine
	c.L.Unlock()

	time.Sleep(100 * time.Millisecond)
}

// ============================================================================
// 3. HOW Cond.Wait() WORKS (THE HIDDEN BEHAVIOR)
// ============================================================================

func condWaitBehavior() {
	fmt.Println("\n=== Understanding Cond.Wait() ===")

	fmt.Println("\nWhen you call c.Wait():")
	fmt.Println("1. Automatically calls c.L.Unlock() (releases the lock)")
	fmt.Println("2. Suspends the goroutine (efficient sleep)")
	fmt.Println("3. Waits for Signal() or Broadcast()")
	fmt.Println("4. Automatically calls c.L.Lock() (re-acquires the lock)")
	fmt.Println("5. Returns (goroutine continues)")

	fmt.Println("\nPattern:")
	fmt.Println("  c.L.Lock()")
	fmt.Println("  for !condition {")
	fmt.Println("      c.Wait() // Unlock -> Wait -> Lock")
	fmt.Println("  }")
	fmt.Println("  c.L.Unlock()")

	fmt.Println("\n  IMPORTANT: Always check condition in a LOOP!")
	fmt.Println("Signal doesn't guarantee your condition is true,")
	fmt.Println("only that SOMETHING happened.")
}

// ============================================================================
// 4. PRODUCER-CONSUMER WITH QUEUE
// ============================================================================

func queueExample() {
	fmt.Println("\n=== Producer-Consumer Queue Example ===")

	// Create condition with a Mutex
	c := sync.NewCond(&sync.Mutex{})

	// Queue with fixed capacity of 2
	queue := make([]interface{}, 0, 10)

	// Consumer: removes items from queue
	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)

		c.L.Lock()        // Enter critical section
		queue = queue[1:] // Dequeue
		fmt.Println("  [Consumer] Removed from queue")
		c.L.Unlock() // Exit critical section

		c.Signal() // Wake up a waiting producer
	}

	// Producer: adds 10 items to queue
	fmt.Println("Adding 10 items to queue (max size: 2)...")
	for i := 0; i < 10; i++ {
		c.L.Lock() // Enter critical section

		// Wait while queue is full
		for len(queue) == 2 {
			fmt.Println("[Producer] Queue full, waiting...")
			c.Wait() // Suspend until consumer signals
		}

		fmt.Printf("[Producer] Adding item %d to queue\n", i+1)
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second) // Start consumer

		c.L.Unlock() // Exit critical section
	}

	fmt.Println("All items added to queue!")
	time.Sleep(2 * time.Second) // Let consumers finish
}

// ============================================================================
// 5. Signal vs Broadcast
// ============================================================================

func signalVsBroadcast() {
	fmt.Println("\n=== Signal vs Broadcast ===")

	fmt.Println("\nSignal():")
	fmt.Println("  • Wakes up ONE goroutine (FIFO order)")
	fmt.Println("  • The one that's been waiting the longest")
	fmt.Println("  • Use when only one goroutine should handle the event")

	fmt.Println("\nBroadcast():")
	fmt.Println("  • Wakes up ALL waiting goroutines")
	fmt.Println("  • Use when all goroutines should handle the event")
	fmt.Println("  • Cannot easily replicate with channels")

	c := sync.NewCond(&sync.Mutex{})
	var wg sync.WaitGroup

	// Start 3 waiting goroutines
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.L.Lock()
			fmt.Printf("  Goroutine %d: Waiting...\n", id)
			c.Wait()
			fmt.Printf("  Goroutine %d: Woken up!\n", id)
			c.L.Unlock()
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nCalling Broadcast() - all goroutines wake up:")
	c.Broadcast() // Wake ALL waiting goroutines

	wg.Wait()
	fmt.Println("All goroutines finished!")
}

// ============================================================================
// 6. BROADCAST EXAMPLE: BUTTON CLICK HANDLERS
// ============================================================================

// Button represents a GUI button with click handlers
type Button struct {
	Clicked *sync.Cond
}

func buttonExample() {
	fmt.Println("\n=== Broadcast Example: Button Click Handlers ===")

	// Create a button with a Cond for click events
	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	// subscribe registers a handler for button clicks
	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunning sync.WaitGroup
		goroutineRunning.Add(1)

		go func() {
			goroutineRunning.Done() // Signal that goroutine started

			c.L.Lock()
			defer c.L.Unlock()

			c.Wait() // Wait for button click
			fn()     // Execute handler
		}()

		goroutineRunning.Wait() // Ensure goroutine started
	}

	// WaitGroup to ensure all handlers execute before exiting
	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)

	// Register 3 different click handlers
	subscribe(button.Clicked, func() {
		fmt.Println("  Handler 1: Maximizing window")
		clickRegistered.Done()
	})

	subscribe(button.Clicked, func() {
		fmt.Println("  Handler 2: Displaying dialog box")
		clickRegistered.Done()
	})

	subscribe(button.Clicked, func() {
		fmt.Println("  Handler 3: Mouse clicked event")
		clickRegistered.Done()
	})

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nSimulating button click...")
	button.Clicked.Broadcast() // Trigger ALL handlers at once!

	clickRegistered.Wait()
	fmt.Println("\nAll handlers executed!")
	fmt.Println("This is hard to do with channels - Cond shines here!")
}

// ============================================================================
// 7. MULTIPLE BROADCASTS
// ============================================================================

func multipleBroadcasts() {
	fmt.Println("\n=== Multiple Broadcasts ===")

	type Button struct {
		Clicked *sync.Cond
	}

	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	// Handler that can be triggered multiple times
	clickHandler := func(id int) {
		for i := 0; i < 3; i++ {
			button.Clicked.L.Lock()
			button.Clicked.Wait()
			fmt.Printf("  Handler %d: Click #%d received\n", id, i+1)
			button.Clicked.L.Unlock()
		}
	}

	// Start 2 handlers
	var wg sync.WaitGroup
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clickHandler(id)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Simulate 3 button clicks
	for click := 1; click <= 3; click++ {
		fmt.Printf("\nClick #%d:\n", click)
		button.Clicked.Broadcast()
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("\nBroadcast can be called multiple times!")
}

// ============================================================================
// 8. WHY THE LOOP? (SPURIOUS WAKEUPS)
// ============================================================================

func whyTheLoop() {
	fmt.Println("\n=== Why Always Check Condition in a Loop? ===")

	fmt.Println("\nReasons to use a loop:")
	fmt.Println("1. Spurious wakeups: Wait() might return without Signal/Broadcast")
	fmt.Println("2. Multiple waiters: Another goroutine might consume the condition")
	fmt.Println("3. Broadcast wakes all: Not all may satisfy the condition")

	fmt.Println("\n WRONG (no loop):")
	fmt.Println("  c.L.Lock()")
	fmt.Println("  if !condition { c.Wait() }")
	fmt.Println("  c.L.Unlock()")

	fmt.Println("\n CORRECT (with loop):")
	fmt.Println("  c.L.Lock()")
	fmt.Println("  for !condition { c.Wait() }")
	fmt.Println("  c.L.Unlock()")

	// Demonstrate multiple waiters
	c := sync.NewCond(&sync.Mutex{})
	ready := false

	// 3 goroutines waiting for the same condition
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			c.L.Lock()
			for !ready { // Loop is essential!
				fmt.Printf("  Goroutine %d: Waiting...\n", id)
				c.Wait()
			}
			fmt.Printf("  Goroutine %d: Condition met!\n", id)
			c.L.Unlock()
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nBroadcasting to all waiters:")
	c.L.Lock()
	ready = true
	c.Broadcast()
	c.L.Unlock()

	wg.Wait()
	fmt.Println("All goroutines checked the condition!")
}

// ============================================================================
// 9. WHEN TO USE Cond
// ============================================================================

func whenToUseCond() {
	fmt.Println("\n=== When to Use Cond ===")

	fmt.Println("\nUse Cond when:")
	fmt.Println("  - You need to broadcast to multiple goroutines")
	fmt.Println("  - Repeated signaling is needed")
	fmt.Println("  - Efficiency is critical (more performant than channels)")
	fmt.Println("  - Waiting for a condition to become true")

	fmt.Println("\nDon't use Cond when:")
	fmt.Println("  - Simple 1-to-1 signaling (use channels)")
	fmt.Println("  - Transferring data (use channels)")
	fmt.Println("  - Select statement needed (use channels)")

	fmt.Println("\nCond vs Channels:")
	fmt.Println("  Cond:     Signaling without data, broadcast capability")
	fmt.Println("  Channels: Transferring data, select statements, easier reasoning")
}

// ============================================================================
// 10. REAL-WORLD: WORKER POOL WITH COND
// ============================================================================

type WorkerPool struct {
	cond     *sync.Cond
	tasks    []string
	mu       sync.Mutex
	shutdown bool
}

func NewWorkerPool() *WorkerPool {
	wp := &WorkerPool{
		tasks: make([]string, 0),
	}
	wp.cond = sync.NewCond(&wp.mu)
	return wp
}

func (wp *WorkerPool) AddTask(task string) {
	wp.cond.L.Lock()
	wp.tasks = append(wp.tasks, task)
	wp.cond.L.Unlock()
	wp.cond.Signal() // Wake up one waiting worker
}

func (wp *WorkerPool) Worker(id int) {
	for {
		wp.cond.L.Lock()

		// Wait for tasks or shutdown
		for len(wp.tasks) == 0 && !wp.shutdown {
			wp.cond.Wait()
		}

		// Check if shutting down
		if wp.shutdown {
			wp.cond.L.Unlock()
			return
		}

		// Get a task
		task := wp.tasks[0]
		wp.tasks = wp.tasks[1:]
		wp.cond.L.Unlock()

		// Process task
		fmt.Printf("  Worker %d: Processing '%s'\n", id, task)
		time.Sleep(100 * time.Millisecond)
	}
}

func (wp *WorkerPool) Shutdown() {
	wp.cond.L.Lock()
	wp.shutdown = true
	wp.cond.L.Unlock()
	wp.cond.Broadcast() // Wake all workers to exit
}

func workerPoolExample() {
	fmt.Println("\n=== Real-World: Worker Pool with Cond ===")

	pool := NewWorkerPool()
	var wg sync.WaitGroup

	// Start 3 workers
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pool.Worker(id)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Add tasks
	fmt.Println("Adding tasks to pool...")
	tasks := []string{"Task A", "Task B", "Task C", "Task D", "Task E"}
	for _, task := range tasks {
		pool.AddTask(task)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\nShutting down workers...")
	pool.Shutdown()
	wg.Wait()
	fmt.Println("All workers shut down!")
}

// ============================================================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================================================

func CondDemo() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              sync.Cond COMPLETE GUIDE                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Run all demonstrations [recommended to run turn by turn]
	inefficientApproaches()
	// basicCond()
	// condWaitBehavior()
	// queueExample()
	// signalVsBroadcast()
	// buttonExample()
	// multipleBroadcasts()
	// whyTheLoop()
	// whenToUseCond()
	// workerPoolExample()

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY TAKEAWAYS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║ sync.Cond = Condition Variable (event signaling)           ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Three Methods:                                             ║")
	fmt.Println("║   • Wait()      - Suspend until signaled                   ║")
	fmt.Println("║   • Signal()    - Wake ONE waiting goroutine               ║")
	fmt.Println("║   • Broadcast() - Wake ALL waiting goroutines              ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Wait() Behavior (IMPORTANT):                               ║")
	fmt.Println("║   1. Unlocks the mutex                                     ║")
	fmt.Println("║   2. Suspends goroutine                                    ║")
	fmt.Println("║   3. Waits for signal                                      ║")
	fmt.Println("║   4. Re-locks the mutex                                    ║")
	fmt.Println("║   5. Returns                                               ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ GOLDEN RULE: Always check condition in a LOOP!             ║")
	fmt.Println("║   c.L.Lock()                                               ║")
	fmt.Println("║   for !condition { c.Wait() }                              ║")
	fmt.Println("║   c.L.Unlock()                                             ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Use Cases:                                                 ║")
	fmt.Println("║   - Broadcasting to multiple goroutines                    ║")
	fmt.Println("║   - Repeated signaling                                     ║")
	fmt.Println("║   - High-performance event notification                    ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║ Broadcast() is the killer feature:                         ║")
	fmt.Println("║   • Hard to replicate with channels                        ║")
	fmt.Println("║   • More performant than channels                          ║")
	fmt.Println("║   • Best constrained to tight scopes                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}
