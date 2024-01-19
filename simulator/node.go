package main

import "sync"

type Node struct {
	id                 int
	ramMemory          int
	currentMs          int
	ramCache           *Cache
	diskCache          *DiskCache
	executingFunctions []*ExecutingFunction
	nodeLock           *sync.Mutex
	MAX_MEMORY         int
}

type Invocation struct {
	hashOwner    string
	hashFunction string
	memory       int
	duration     int
	timestamp    int
}

type ExecutingFunction struct {
	function string
	memory   int
	end      int
}

func createNode(id int, memoryRAM int) Node {
	var n Node
	n.id = id
	n.ramMemory = memoryRAM
	n.MAX_MEMORY = memoryRAM
	n.currentMs = 0
	n.diskCache = createDisk(props.diskMemory)
	n.ramCache = createCache(n.diskCache)
	n.executingFunctions = make([]*ExecutingFunction, 0)
	n.nodeLock = new(sync.Mutex)
	return n
}

func minToMs(minutes int) int {
	return minutes * 60 * 1000
}

func updateNode(node *Node, ms int) {

	updateCache(node, node.ramCache, ms)
	updateDisk(node, node.diskCache, ms)

	i := 0
	for ; i < len(node.executingFunctions); i++ {
		if node.executingFunctions[i].end > ms {
			break
		} else {
			item := node.executingFunctions[i]
			insertCacheItem(node.ramCache, item.function, item.memory, item.end)
		}
	}
	node.executingFunctions = node.executingFunctions[i:]
	node.currentMs = ms

}

func updateNodes(nodeList *[]Node, ms int, stats *Statistics) {

	for i := 0; i < props.nNodes; i++ {
		(*nodeList)[i].nodeLock.Lock()
		updateNode(&(*nodeList)[i], ms)

		stats.statsLock.Lock()
		if ms-stats.statsMs >= 1000 {
			diskMem := props.diskMemory - (*nodeList)[i].diskCache.diskMemory
			ramMem := props.ramMemory - (*nodeList)[i].ramMemory

			if i == N_NODES-1 {
				writeStats(stats, 0, ramMem, ms/1000, diskMem)
				stats.statsMs = ms
			}
		}
		stats.statsLock.Unlock()

		(*nodeList)[i].nodeLock.Unlock()
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

	updateNode(node, invocation.timestamp)

	latency := 0

	//Search for function in RAMcache
	inCache := searchCache(node.ramCache, invocation.hashFunction)

	//If in cache, sign the cache to remove it
	if inCache {
		stats.warmStarts[node.id]++
		stats.statsLock.Lock()
		stats.warmStartsSecond++
		stats.statsLock.Unlock()
		retrieveCache(node.ramCache, invocation.hashFunction)
	} else {

		if node.ramMemory < invocation.memory {
			freedMem := freeCache(node.ramCache, invocation.memory-node.ramMemory)
			node.ramMemory += freedMem
			if node.ramMemory < invocation.memory {
				freed := freeBuffer(node.diskCache, invocation.memory-node.ramMemory)
				node.ramMemory += freed
				node.ramCache.occupied -= freed
				if node.ramMemory < invocation.memory {
					stats.failedInvocations[node.id]++
					node.nodeLock.Unlock()
					stats.statsLock.Lock()
					stats.failedInvocationsSecond++
					stats.statsLock.Unlock()
					return
				}
			}
		}

		inDisk := searchDisk(node.diskCache, invocation.hashFunction)

		if inDisk {
			latency = addToReadQueue(node.diskCache, invocation.hashFunction, invocation.memory, invocation.timestamp)

			if latency >= 0 {
				stats.lukewarmStarts[node.id]++
				stats.statsLock.Lock()
				stats.lukeWarmStartsSecond++
				stats.statsLock.Unlock()
			} else {
				stats.coldStarts[node.id]++
				stats.statsLock.Lock()
				stats.coldStartsSecond++
				stats.statsLock.Unlock()
				latency = props.coldLatency
			}

		} else {
			stats.coldStarts[node.id]++
			stats.statsLock.Lock()
			stats.coldStartsSecond++
			stats.statsLock.Unlock()
			latency = props.coldLatency
		}
		node.ramMemory -= invocation.memory
	}

	if stats.invocations[node.id]%1000 == 0 {
		stats.latencyCdf[node.id] = append(stats.latencyCdf[node.id], latency)
	}

	//Add it to the executing functions
	addToExecuting(node, invocation.hashFunction, invocation.timestamp+invocation.duration+latency, invocation.memory)

	node.nodeLock.Unlock()

}

func findNode(nodeList *[]Node, ms int, stats *Statistics, function string, memory int) int {
	updateNodes(nodeList, ms, stats)
	i := 0
	chosenNode := 0
	chosenWarm := false
	chosenLukewarm := false
	for ; i < props.nNodes; i++ {
		(*nodeList)[i].nodeLock.Lock()

		_, existsWarm := (*nodeList)[i].ramCache.functionMap[function]
		if existsWarm {
			if !chosenWarm {
				chosenNode = i
				chosenWarm = true
			} else {
				if (*nodeList)[i].ramMemory > (*nodeList)[chosenNode].ramMemory {
					chosenNode = i
				}
			}
			(*nodeList)[i].nodeLock.Unlock()
			continue
		}

		if chosenWarm {
			(*nodeList)[i].nodeLock.Unlock()
			continue
		}

		_, existsLukewarm := (*nodeList)[i].diskCache.functionMap[function]
		if (*nodeList)[i].ramMemory+(*nodeList)[i].ramCache.occupied < memory {
			(*nodeList)[i].nodeLock.Unlock()
			continue
		}
		if existsLukewarm {
			if !chosenLukewarm {
				chosenNode = i
				chosenLukewarm = true
			} else {
				if (*nodeList)[i].ramMemory > (*nodeList)[chosenNode].ramMemory {
					chosenNode = i
				}
			}
			(*nodeList)[i].nodeLock.Unlock()
			continue
		}

		if chosenLukewarm {
			(*nodeList)[i].nodeLock.Unlock()
			continue
		}

		if (*nodeList)[i].ramMemory > memory {
			if (*nodeList)[i].ramMemory > (*nodeList)[chosenNode].ramMemory {
				chosenNode = i
			}
		}

		(*nodeList)[i].nodeLock.Unlock()
	}
	return chosenNode
}
