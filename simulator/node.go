package main

import "sync"

type Node struct {
	id                 int
	runMemory          int
	ramMemory          int
	currentMs          int
	ramCache           *RAMCache
	executingFunctions []*ExecutingFunction
	nodeLock           *sync.Mutex
}

type ExecutingFunction struct {
	function string
	memory   int
	end      int
}

func createNode(id int, memory int, memoryRAM int) Node {
	var n Node
	n.id = id
	n.runMemory = memory
	n.currentMs = 0
	n.ramMemory = memoryRAM
	n.ramCache = createRAMCache(memoryRAM)
	n.executingFunctions = make([]*ExecutingFunction, 0)
	n.nodeLock = new(sync.Mutex)
	return n
}

func minToMs(minutes int) int {
	return minutes * 60 * 1000
}

func updateNode(node *Node, ms int) {

	node.nodeLock.Lock()

	updateRAMCache(node.ramCache, ms)

	i := 0
	for ; i < len(node.executingFunctions); i++ {
		if node.executingFunctions[i].end > ms {
			break
		} else {
			item := node.executingFunctions[i]
			node.runMemory += item.memory
			insertRAMItem(node.ramCache, item.function, item.memory, item.end+minToMs(KEEP_ALIVE_WINDOW))
		}
	}
	node.executingFunctions = node.executingFunctions[i:]
	node.currentMs = ms

	node.nodeLock.Unlock()

}

func updateNodes(nodeList *[N_NODES]Node, ms int) {
	for i := range nodeList {
		updateNode(&nodeList[i], ms)
	}
}

func addToExecuting(node *Node, function string, end int, memory int) {
	i := 0
	for ; i < len(node.executingFunctions); i++ {
		if node.executingFunctions[i].end > end {
			break
		}
	}
	newFunction := new(ExecutingFunction)
	newFunction.function = function
	newFunction.memory = memory
	newFunction.end = end
	node.executingFunctions = append(node.executingFunctions, nil)
	copy(node.executingFunctions[i+1:], node.executingFunctions[i:])
	node.executingFunctions[i] = newFunction
}

func allocateInvocation(node *Node, invocation Invocation, stats *Statistics) {

	node.nodeLock.Lock()

	if node.runMemory < invocation.memory {
		stats.failedInvocations[node.id]++
		node.nodeLock.Unlock()
		return
	}

	//Search for function in RAMcache
	inCache := searchRAMCache(node.ramCache, invocation.hashFunction)

	//If in cache, sign the cache to remove it
	if inCache {
		stats.warmStarts[node.id]++
		retrieveRAMCache(node.ramCache, invocation.hashFunction)
	} else {
		stats.coldStarts[node.id]++
	}

	node.runMemory -= invocation.memory

	//Add it to the executing functions
	addToExecuting(node, invocation.hashFunction, invocation.timestamp+invocation.duration, invocation.memory)

	node.nodeLock.Unlock()

}

func findNode(nodeList *[N_NODES]Node, ms int) int {
	updateNodes(nodeList, ms)
	i := 0
	chosenNode := 0
	auxMemory := -1
	for ; i < len(nodeList); i++ {
		nodeList[i].nodeLock.Lock()
		if nodeList[i].runMemory > auxMemory {
			auxMemory = nodeList[i].runMemory
			chosenNode = i
		}
		nodeList[i].nodeLock.Unlock()
	}
	return chosenNode
}
