package main

import (
	"fmt"
	"time"
)

// =============================================================================
// TOPIC: What Is CSP? (Communicating Sequential Processes)
// =============================================================================
// CSP stands for "Communicating Sequential Processes"
// - Introduced by Tony Hoare in 1978
// - Core idea: Input and output are overlooked primitives in concurrent programming
// - Go's concurrency model is inspired by CSP principles
//
// KEY CONCEPTS:
// 1. Processes: Encapsulated logic that requires input and produces output
// 2. Communication: Processes communicate via channels (not shared memory)
// 3. Composition: Channels are composable, making large systems simpler
//
// HOARE'S ORIGINAL NOTATION:
// - ! (send): Send input INTO a process
// - ? (read): Read output FROM a process
//
// GO'S EQUIVALENT:
// - channel <- value  (send)
// - value := <-channel (receive)
// =============================================================================

// Example 1: Basic Channel Communication
// Demonstrates: Process A sends to Process B
func BasicChannelDemo() {
	fmt.Println("=== Basic Channel Communication ===")

	// Create a channel (the communication primitive)
	messages := make(chan string)

	// Process A: Producer (sends output)
	go func() {
		messages <- "Hello from Process A, How you doing"
	}()

	// Process B: Consumer (reads input)
	msg := <-messages
	fmt.Println("Process B received:", msg)
	fmt.Println()
}

// Example 2: Corresponding Processes
// Demonstrates: Output from one process flows directly into input of another
// This is what Hoare called "correspondence"
func CorrespondingProcesses() {
	fmt.Println("=== Corresponding Processes ===")

	// west -> c -> east (from Hoare's paper example)
	west := make(chan rune) // rune is an alias for int32 and is equivalent to int32 in all ways. [copied from official docs]
	east := make(chan rune)

	// West process: produces characters
	go func() {
		message := "CSP"
		for _, char := range message {
			west <- char
		}
		close(west)
	}()

	// Middle process: reads from west, sends to east
	// *[c:character; west?c → east!c]
	go func() {
		for c := range west {
			east <- c
		}
		close(east)
	}()

	// East process: consumes characters
	fmt.Print("East received: ")
	for c := range east {
		fmt.Printf("%c", c)
	}
	fmt.Println("")
}

// Example 3: Select Statement (Go's Enhancement)
// Demonstrates: Guarded commands - conditional execution based on channel state
func SelectStatementDemo() {
	fmt.Println("=== Select Statement (Guarded Commands) ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	// Two competing processes
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "from channel 1"
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch2 <- "from channel 2, HAHAHHAHAHA keep sleeping rabbit"
	}()

	// Select waits for whichever channel is ready first
	// This is Go's implementation of guarded commands
	select {
	case msg1 := <-ch1:
		fmt.Println("Received", msg1)
	case msg2 := <-ch2:
		fmt.Println("Received", msg2)
	}
	fmt.Println()
}

// Example 4: Channel Composition
// Demonstrates: Combining multiple channels to coordinate subsystems
func ChannelComposition() {
	fmt.Println("=== Channel Composition ===")

	// Three subsystems producing data
	source1 := make(chan int)
	source2 := make(chan int)
	source3 := make(chan int)

	// Combined output
	combined := make(chan int)

	// Producer subsystems
	go func() {
		source1 <- 1
		close(source1)
	}()
	go func() {
		source2 <- 2
		close(source2)
	}()
	go func() {
		source3 <- 3
		close(source3)
	}()

	// Coordinator: composes inputs from multiple sources
	go func() {
		combined <- <-source1
		combined <- <-source2
		combined <- <-source3
		close(combined)
	}()

	// Consumer
	fmt.Print("Combined output: ")
	for val := range combined {
		fmt.Printf("%d ", val)
	}
	fmt.Println("")
}

// Example 5: Web Server Pattern (From the book)
// Demonstrates: Natural mapping of concurrent problems to Go code
func WebServerPattern() {
	fmt.Println("=== Web Server Pattern ===")
	fmt.Println("Natural concurrency: One goroutine per connection")

	// Simulate incoming connections
	connections := make(chan int)

	// Connection handler (one goroutine per user)
	handleConnection := func(connID int) {
		fmt.Printf("Handling connection %d\n", connID)
		time.Sleep(50 * time.Millisecond) // Simulate work
		fmt.Printf("Connection %d complete\n", connID)
	}

	// Spawn handler goroutines (no thread pool needed!)
	go func() {
		for i := 1; i <= 3; i++ {
			connections <- i
		}
		close(connections)
	}()

	// Natural problem mapping: one goroutine per connection
	for connID := range connections {
		go handleConnection(connID)
	}

	time.Sleep(200 * time.Millisecond) // Wait for handlers
	fmt.Println()
}

// Example 6: Timeout Pattern
// Demonstrates: Composing channels with time (cancellation/timeout)
func TimeoutPattern() {
	fmt.Println("=== Timeout Pattern ===")

	slowProcess := make(chan string)

	go func() {
		time.Sleep(2 * time.Second)
		slowProcess <- "finally done"
	}()

	select {
	case result := <-slowProcess:
		fmt.Println("Received:", result)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("Timeout! Process took too long")
	}
	fmt.Println()
}

func CspBasics() {
	fmt.Println("CSP (Communicating Sequential Processes) Demonstrations")
	fmt.Println("========================================================")

	BasicChannelDemo()
	CorrespondingProcesses()
	SelectStatementDemo()
	ChannelComposition()
	WebServerPattern()
	TimeoutPattern()

	fmt.Println("Key Takeaways:")
	fmt.Println("1. Channels = communication primitives (not shared memory)")
	fmt.Println("2. Goroutines = lightweight processes")
	fmt.Println("3. Select = guarded commands for channel coordination")
	fmt.Println("4. Composition = building large systems from small pieces")
}
