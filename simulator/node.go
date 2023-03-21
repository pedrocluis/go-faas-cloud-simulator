package main

import "sync"

const KEEP_ALIVE_WINDOW = 5
const MAX_MINUTES = 10000

type Node struct {
	memory                   int
	availableMemoryPerMinute []int
	appsInMemoryPerMinute    []map[string]*AppInMemory
}

type AppInMemory struct {
	memory     int
	containers int
}

// Create a node with a memory capacity
func newNode(memory int) Node {
	var n Node
	n.memory = memory

	//Initialize available memory per minute array
	n.availableMemoryPerMinute = make([]int, MAX_MINUTES+1)
	n.appsInMemoryPerMinute = make([]map[string]*AppInMemory, MAX_MINUTES+1)
	for i := range n.availableMemoryPerMinute {
		n.availableMemoryPerMinute[i] = memory
	}
	return n
}

// Allocate memory for an app for each minute one of its functions is being used and for the keep-alive
func allocateMemory(node *Node, app string, minute int, memory int, duration int, coldStarts *int, coldStartsLock *sync.Mutex) {

	// Check to see if app is loaded
	var i int
	inMemory := false
	// Check for all the containers where their keep alive windows catches this function's invocation
	for i = minute + 1 - KEEP_ALIVE_WINDOW; i <= minute; i++ {

		//The first minute is minute 1
		if i < 1 {
			continue
		}

		//Initialize the map
		if node.appsInMemoryPerMinute[i] == nil {
			node.appsInMemoryPerMinute[i] = make(map[string]*AppInMemory)
		}

		_, contains := node.appsInMemoryPerMinute[i][app]
		if contains {
			//Occupy a container
			node.appsInMemoryPerMinute[i][app].containers--
			if node.appsInMemoryPerMinute[i][app].containers == 0 {
				delete(node.appsInMemoryPerMinute[i], app)
			}
			inMemory = true
			break
		}
	}

	//If the app is in memory, occupy the rest of the memory from when the container was scheduled to be unloaded to the function's end
	if inMemory {
		for i = i + KEEP_ALIVE_WINDOW - 1; i < minute+duration; i++ {
			node.availableMemoryPerMinute[i] -= memory
		}
	} else {
		//If it's not in memory, occupy the memory from the beginning of the function until it's end
		for i = minute; i < minute+duration; i++ {
			node.availableMemoryPerMinute[i] -= memory
		}
		coldStartsLock.Lock()
		*coldStarts++
		coldStartsLock.Unlock()
	}

	// Keep the app loaded in memory starting at the function's end
	// Initialize the map
	if node.appsInMemoryPerMinute[minute+duration] == nil {
		node.appsInMemoryPerMinute[minute+duration] = make(map[string]*AppInMemory)
	}

	_, contains := node.appsInMemoryPerMinute[minute+duration][app]
	if !contains {
		//If the app has never been loaded create the key for it in the map
		newElement := AppInMemory{memory: memory, containers: 1}
		node.appsInMemoryPerMinute[minute+duration][app] = &newElement
	} else {
		//Add one more free container
		node.appsInMemoryPerMinute[minute+duration][app].containers++
	}
	//Occupy the memory for the keep-alive period
	for i := minute + duration; i < minute+duration+KEEP_ALIVE_WINDOW; i++ {
		node.availableMemoryPerMinute[i] -= memory
	}
}

// Search for containers with an app loaded that are not in use
func unloadMemory(minute int, memory int, node *Node) bool {
	freedMemory := 0
	//Check all the containers with loaded apps that are not in use and reach our function invocation
	for i := minute - KEEP_ALIVE_WINDOW + 1; i <= minute; i++ {
		for app := range node.appsInMemoryPerMinute[i] {
			numContainers := node.appsInMemoryPerMinute[i][app].containers
			//Unload the containers one by one until we've freed the memory necessary
			for containers := 0; containers < numContainers; containers++ {
				node.appsInMemoryPerMinute[i][app].containers--
				node.availableMemoryPerMinute[i] -= node.appsInMemoryPerMinute[i][app].memory
				freedMemory += node.appsInMemoryPerMinute[i][app].memory
				if node.appsInMemoryPerMinute[i][app].containers == 0 {
					delete(node.appsInMemoryPerMinute[i], app)
				}
				if freedMemory >= memory {
					return true
				}
			}
		}
	}
	return false
}
