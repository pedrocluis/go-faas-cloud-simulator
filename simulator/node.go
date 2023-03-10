package main

import "golang.org/x/exp/slices"

const KEEP_ALIVE_WINDOW = 10

type Node struct {
	memory                   int
	availableMemoryPerMinute []int
	appsInMemoryPerMinute    [][]string
}

func newNode(memory int) Node {
	var n Node
	n.memory = memory

	//Initialize available memory per minute array
	n.availableMemoryPerMinute = make([]int, 1441)
	n.appsInMemoryPerMinute = make([][]string, 1441)
	for i := range n.availableMemoryPerMinute {
		n.availableMemoryPerMinute[i] = memory
	}
	return n
}

func allocateMemory(node Node, app string, minute int, memory int, duration int, coldStarts *int) {

	for i := minute; i < minute+duration+KEEP_ALIVE_WINDOW; i++ {
		if !slices.Contains(node.appsInMemoryPerMinute[i], app) {
			// At the minute the function is called, if it's not in memory, it's a cold start
			if i == minute {
				*coldStarts++
			}
			node.availableMemoryPerMinute[i] -= memory
			node.appsInMemoryPerMinute[i] = append(node.appsInMemoryPerMinute[i], app)
		}
	}
}
