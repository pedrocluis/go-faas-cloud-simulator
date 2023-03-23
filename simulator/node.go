package main

import (
	"sort"
	"sync"
)

const KEEP_ALIVE_WINDOW = 5 // Keep alive in minutes
const MINUTES_IN_DAY = 1440 //We add 2000 extra

type Node struct {
	memory                        int
	availableMemoryPerMillisecond []int
	appsInMemory                  map[string]*ContainersInMemory
	orderedContainers             []*OrderedContainers
}

type ContainersInMemory struct {
	memory             int
	containerStartTime []int
}

type OrderedContainers struct {
	app string
	ms  int
}

func minToMs(minutes int) int {
	return minutes * 60 * 1000
}

// Create a node with a memory capacity
func newNode(memory int) Node {
	var n Node
	n.memory = memory

	//Initialize available memory per minute array
	n.availableMemoryPerMillisecond = make([]int, minToMs(MINUTES_IN_DAY+1))
	n.appsInMemory = make(map[string]*ContainersInMemory)
	for i := range n.availableMemoryPerMillisecond {
		n.availableMemoryPerMillisecond[i] = memory
	}
	n.orderedContainers = make([]*OrderedContainers, 0)
	return n
}

func updateOrderedContainers(node *Node, ms int) {
	//Update the containers to only include the ones where their keep-alive catches this invocation
	i := sort.Search(len(node.orderedContainers),
		func(i int) bool { return node.orderedContainers[i].ms >= ms-minToMs(KEEP_ALIVE_WINDOW) })

	if i < len(node.orderedContainers) {
		node.orderedContainers = node.orderedContainers[i:]
	}
}

func updateAppContainers(node *Node, app string, ms int) {
	//Update the containers to only include the ones where their keep-alive catches this invocation
	i := sort.Search(len(node.appsInMemory[app].containerStartTime),
		func(i int) bool { return node.appsInMemory[app].containerStartTime[i] >= ms-minToMs(KEEP_ALIVE_WINDOW) })
	node.appsInMemory[app].containerStartTime = node.appsInMemory[app].containerStartTime[i:]
	//If no container left, remove app from hashtable
	if len(node.appsInMemory[app].containerStartTime) == 0 {
		delete(node.appsInMemory, app)
	}
}

func insertOrderedContainers(ordered []*OrderedContainers, element *OrderedContainers) []*OrderedContainers {
	var dummy *OrderedContainers
	i := sort.Search(len(ordered), func(i int) bool { return ordered[i].ms >= element.ms })
	ordered = append(ordered, dummy)
	copy(ordered[i+1:], ordered[i:])
	ordered[i] = element
	return ordered
}

func insertOrderedApp(ordered []int, start int) []int {
	var dummy int
	i := sort.Search(len(ordered), func(i int) bool { return ordered[i] >= start })
	ordered = append(ordered, dummy)
	copy(ordered[i+1:], ordered[i:])
	ordered[i] = start
	return ordered
}

func removeOrderedContainers(ordered []*OrderedContainers, appName string) []*OrderedContainers {
	var idx int
	for idx = range ordered {
		if ordered[idx].app == appName {
			break
		}
	}
	return append(ordered[:idx], ordered[idx+1:]...)
}

// Allocate memory for an app for each minute one of its functions is being used and for the keep-alive
func allocateMemory(node *Node, app string, minute int, memory int, duration int, coldStarts *int, coldStartsLock *sync.Mutex, failedInvocations *int, failLock *sync.Mutex) {

	//Also update the ordered list
	updateOrderedContainers(node, minToMs(minute))

	// Variable to know if we caught a free container and if we did, which ms does it start at
	caughtContainerMs := -1

	// Check to see if the app is loaded
	_, contains := node.appsInMemory[app]
	if contains {

		//Update the app container's list
		updateAppContainers(node, app, minToMs(minute))

		//Check if the app still exists in the hash table
		_, containsAgain := node.appsInMemory[app]
		if containsAgain {
			if node.appsInMemory[app].containerStartTime[0] <= minToMs(minute) {
				caughtContainerMs = node.appsInMemory[app].containerStartTime[0]
				// Try to reserve the memory
				for ms := caughtContainerMs; ms < minToMs(minute)+duration; ms++ {
					if node.availableMemoryPerMillisecond[ms] < memory {
						if !unloadMemory(ms, memory, node, app) {
							caughtContainerMs = -1
							break
						}
					}
				}
			}
		}
	}

	//If the app is in memory, occupy the rest of the memory from when the container was scheduled to be unloaded to the function's end
	if caughtContainerMs != -1 {
		//If we enter this branch, we have already freed enough memory for the function duration
		// We have to occupy memory enough from where the container was scheduled to end to the function end
		for i := caughtContainerMs + minToMs(KEEP_ALIVE_WINDOW-1); i < minToMs(minute)+duration; i++ {
			node.availableMemoryPerMillisecond[i] -= memory
		}
		//We also have to "occupy the containers"
		node.orderedContainers = removeOrderedContainers(node.orderedContainers, app)
		node.appsInMemory[app].containerStartTime = node.appsInMemory[app].containerStartTime[1:]
		if len(node.appsInMemory[app].containerStartTime) == 0 {
			delete(node.appsInMemory, app)
		}

	} else {
		//If it's not in memory, or we can't use a container due to memory, occupy the memory from the beginning of the function until it's end
		//Check to see if we have available memory
		for i := minToMs(minute); i < minToMs(minute)+duration; i++ {
			if node.availableMemoryPerMillisecond[i] < memory {
				if !unloadMemory(i, memory, node, app) {
					// If we can't unload the necessary memory, do something
					failLock.Lock()
					*failedInvocations++
					failLock.Unlock()
					return
				}
			}
		}
		for i := minToMs(minute); i < minToMs(minute)+duration; i++ {
			node.availableMemoryPerMillisecond[i] -= memory
		}
		coldStartsLock.Lock()
		*coldStarts++
		coldStartsLock.Unlock()
	}

	// Keep the app loaded in memory starting at the function's end
	//Check to see if we can keep the container with the app loaded in a container for the keep-alive window
	for i := minToMs(minute) + duration; i < minToMs(minute)+duration+minToMs(KEEP_ALIVE_WINDOW); i++ {
		if node.availableMemoryPerMillisecond[i] < memory {
			if !unloadMemory(i, memory, node, app) {
				// If there's a millisecond in the keep-alive where we can't find the memory, don't keep the container
				return
			}
		}
	}
	_, contains = node.appsInMemory[app]
	if !contains {
		//If the app has never been loaded create the key for it in the map
		newElement := ContainersInMemory{memory: memory, containerStartTime: []int{minToMs(minute) + duration}}
		node.appsInMemory[app] = &newElement
	} else {
		//Add one more free container with the app loaded (or extend)
		node.appsInMemory[app].containerStartTime = insertOrderedApp(node.appsInMemory[app].containerStartTime, minToMs(minute)+duration)
	}
	//Occupy the memory for the keep-alive period
	for i := minToMs(minute) + duration; i < minToMs(minute)+duration+minToMs(KEEP_ALIVE_WINDOW); i++ {
		node.availableMemoryPerMillisecond[i] -= memory
	}
	node.orderedContainers = insertOrderedContainers(node.orderedContainers, &OrderedContainers{app: app, ms: minToMs(minute) + duration})

}

// Search for containers with an app loaded that are not in use
func unloadMemory(millisecond int, memory int, node *Node, appName string) bool {

	freedMemory := 0
	var undoContainer *OrderedContainers = nil
	sameApp := false
	for {

		if len(node.orderedContainers) == 0 {
			if undoContainer != nil {
				//If we deleted the container we are going to use, put it back
				node.orderedContainers = undoDeleteOrdered(node.orderedContainers, undoContainer)
				_, contains := node.appsInMemory[undoContainer.app]
				if !contains {
					node.appsInMemory[undoContainer.app] = &ContainersInMemory{memory: memory, containerStartTime: []int{undoContainer.ms}}
				} else {
					node.appsInMemory[undoContainer.app].containerStartTime = undoDeleteApp(node.appsInMemory[undoContainer.app].containerStartTime, undoContainer.ms)
				}
			}
			return false
		}
		app := node.orderedContainers[0].app
		// If the container we're looking at is for the app we're trying to make memory for, we'll have to readd it
		if app == appName {
			if sameApp == false {
				undoContainer = node.orderedContainers[0]
				sameApp = true
				node.orderedContainers = node.orderedContainers[1:]
				node.appsInMemory[app].containerStartTime = node.appsInMemory[app].containerStartTime[1:]
				if len(node.appsInMemory[app].containerStartTime) == 0 {
					delete(node.appsInMemory, app)
				}
				continue
			}
		}
		start := node.orderedContainers[0].ms
		if start >= millisecond {
			return false
		}
		freedMemory += node.appsInMemory[app].memory
		for i := start; i < start+KEEP_ALIVE_WINDOW; i++ {
			node.availableMemoryPerMillisecond[i] -= node.appsInMemory[app].memory
		}
		node.orderedContainers = node.orderedContainers[1:]
		node.appsInMemory[app].containerStartTime = node.appsInMemory[app].containerStartTime[1:]
		if len(node.appsInMemory[app].containerStartTime) == 0 {
			delete(node.appsInMemory, app)
		}
		if freedMemory >= memory {
			if undoContainer != nil {
				//If we deleted the container we're going to use, put it back
				node.orderedContainers = undoDeleteOrdered(node.orderedContainers, undoContainer)
				_, contains := node.appsInMemory[undoContainer.app]
				if !contains {
					node.appsInMemory[undoContainer.app] = &ContainersInMemory{memory: memory, containerStartTime: []int{undoContainer.ms}}
				} else {
					node.appsInMemory[undoContainer.app].containerStartTime = undoDeleteApp(node.appsInMemory[undoContainer.app].containerStartTime, undoContainer.ms)
				}
			}
			return true
		}
	}
}

func undoDeleteOrdered(orderedContainers []*OrderedContainers, app *OrderedContainers) []*OrderedContainers {
	var dummy *OrderedContainers
	orderedContainers = append(orderedContainers, dummy)
	copy(orderedContainers[1:], orderedContainers)
	orderedContainers[0] = app
	return orderedContainers
}

func undoDeleteApp(starts []int, undo int) []int {
	var dummy = 0
	starts = append(starts, dummy)
	copy(starts[1:], starts)
	starts[0] = undo
	return starts
}
