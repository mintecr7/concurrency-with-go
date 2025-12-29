# Concurrency in Go - Learning Repository

A structured learning repository following the book "Concurrency in Go" by Katherine Cox-Buday. This repository contains code examples, exercises, and notes as I explore Go's concurrency primitives and patterns.

## About

This repository documents my journey learning concurrent programming in Go. It covers fundamental concepts like goroutines and channels, as well as advanced patterns for building reliable concurrent systems.

## Repository Structure

```
.
├── ch01-introduction/          # Introduction to concurrency concepts
├── ch02-goroutines/            # Goroutines and the Go runtime
├── ch03-sync-primitives/       # sync package primitives (WaitGroups, Mutex, etc.)
├── ch04-channels/              # Channels and channel operations
├── ch05-concurrency-patterns/  # Common concurrency patterns
├── ch06-runtime/               # Go runtime and concurrency
└── examples/                   # Additional practice examples
```

## Topics Covered

- **Goroutines**: Lightweight concurrent functions
- **Channels**: Communication between goroutines
- **Select Statement**: Multiplexing channel operations
- **Sync Package**: Mutexes, WaitGroups, Once, Pool, Cond
- **Concurrency Patterns**: 
  - Pipeline pattern
  - Fan-out, fan-in
  - Or-channel pattern
  - Context package usage
  - Error handling in concurrent code
- **Concurrency at Scale**: Managing goroutine lifecycles and preventing leaks

## Running the Examples

Each chapter directory contains standalone Go programs. To run an example:

```bash
cd ch04-channels
go run example-name.go
```

Some examples may include tests:

```bash
go test ./...
```

## Prerequisites

- Go 1.21 or higher
- Basic understanding of Go syntax and fundamentals

## Learning Resources

- **Primary Resource**: "Concurrency in Go" by Katherine Cox-Buday (O'Reilly)
- [Go Concurrency Patterns (Go Blog)](https://go.dev/blog/pipelines)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)

## Notes

This is a learning repository, so code may contain experiments, incomplete implementations, and comments explaining concepts. The goal is understanding, not production-ready code.

## Progress Tracker

**:white_check_mark:** Chapter 1: An Introduction to Concurrency
- [ ] Chapter 2: Modeling Your Code: Communicating Sequential Processes
- [ ] Chapter 3: Go's Concurrency Building Blocks
- [ ] Chapter 4: Concurrency Patterns in Go
- [ ] Chapter 5: Concurrency at Scale
- [ ] Chapter 6: Goroutines and the Go Runtime

## License

This repository is for educational purposes. Code examples are based on or inspired by "Concurrency in Go" by Katherine Cox-Buday.
