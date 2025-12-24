package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Livelock :- A state where concurrent processes are constantly changing state
// in response to each other, but none are able to make forward progress.
//
// ANALOGY: Two people meeting in a narrow hallway. Both move to the side to let
// the other pass, but they move to the SAME side at the SAME time,
// blocking each other again. They repeat this indefinitely.

var cadence = sync.NewCond(&sync.Mutex{})

func init() {
	// This goroutine simulates the "beat" of the world.
	// Every millisecond, it tells everyone they can try to take a step.
	go func() {
		for range time.Tick(1 * time.Millisecond) {
			cadence.Broadcast()
		}
	}()
}

func takeStep() {
	cadence.L.Lock()
	cadence.Wait() // Wait for the next "beat" from the broadcaster
	cadence.L.Unlock()
}

// tryDir attempts to move in a specific direction.
// 1. We declare our intent to move in this direction (atomic increment).
// 2. We wait a beat (takeStep) to see if anyone else tried to move here.
// 3. If the count is 1, we are the only ones here—Success!
// 4. If the count > 1, someone else is blocking us. We "politely" step back (decrement).
func tryDir(dirName string, dir *int32, out *bytes.Buffer) bool {
	fmt.Fprintf(out, " %v", dirName)
	atomic.AddInt32(dir, 1) // 1. Declare intent
	takeStep()              // 2. Synchronize cadence

	if atomic.LoadInt32(dir) == 1 { // 3. Check if path is clear
		fmt.Fprint(out, ". Success!")
		return true
	}

	takeStep()
	atomic.AddInt32(dir, -1) // 4. Path blocked, give up and revert state
	fmt.Fprint(out, ". Blocked!")
	return false
}

func runLivelock() {
	var wg sync.WaitGroup
	var left, right int32

	// Helper function for a person walking in the hallway
	walk := func(name string) {
		var out bytes.Buffer
		defer wg.Done()
		defer func() { fmt.Println(out.String()) }()

		fmt.Fprintf(&out, "%v is trying to scoot:", name)

		// We limit to 5 attempts so the program actually finishes.
		// In a real livelock, this loop would go on forever.
		for i := 0; i < 5; i++ {
			if tryDir("left", &left, &out) || tryDir("right", &right, &out) {
				return
			}
		}
		fmt.Fprintf(&out, "\n%v tosses her hands up in exasperation!", name)
	}

	wg.Add(2)
	go walk("Alice")
	go walk("Barbara")
	wg.Wait()
}

func main() {
	runLivelock()
}

// --- What is happening here? ---
//
// 1. ACTIVE WAITING: Unlike Deadlock, where goroutines are suspended,
//    here Alice and Barbara are actively executing code and consuming CPU.
//
// 2. THE SYNC PROBLEM: Because they move at the exact same "cadence,"
//    they both pick 'left' at the same time, see it's blocked,
//    then both pick 'right' at the same time, and see it's blocked.
//
// 3. LACK OF COORDINATION: They are trying to avoid a collision (deadlock)
//    but because their logic is identical and perfectly synchronized,
//    they keep repeating the same failing state.
//
// 4. DETECTION: Livelocks are harder to find than deadlocks. Monitoring
//    tools will show the CPU is busy and the process is "running," but
//    the business logic (getting to the end of the hallway) never completes.
