package main

import (
	syncpackage "learning-concurrency/ch03_go_concurrency_building_blocks/sync_package"
)

func main() {
	// goRoutine()
	// syncpackage.WaitGroupDemo()
	syncpackage.MutexAndRWMutex()
	// syncpackage.CondDemo()
}
