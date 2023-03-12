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

	for i := minute; i <= minute+duration+KEEP_ALIVE_WINDOW; i++ {
		if i > 1440 {
			break
		}
		if !slices.Contains(node.appsInMemoryPerMinute[i], app) {
			// At the minute the function is called, if it's not in memory, allocate the memory
			node.availableMemoryPerMinute[i] -= memory
			node.appsInMemoryPerMinute[i] = append(node.appsInMemoryPerMinute[i], app)
		}
	}
}

func countColdStarts(n Node) int {
	coldStarts := 0
	for i := 1; i <= 1440; i++ {
		for j := range n.appsInMemoryPerMinute[i] {
			if slices.Contains(n.appsInMemoryPerMinute[i-1], n.appsInMemoryPerMinute[i][j]) {
				continue
			}
			coldStarts++
		}
	}
	return coldStarts
}
