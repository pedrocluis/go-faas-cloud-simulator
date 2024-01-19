package main

type FunctionInCache struct {
	name   string
	memory int
	copies int
}

type CacheItem struct {
	function string
	end      int
}

type QueueItem struct {
	function    string
	memory      int
	transferEnd int
}

type Cache struct {
	functionMap      map[string]*FunctionInCache
	orderedFunctions []*CacheItem
	diskMemory       int
	isRam            bool
	occupied         int
	destCache        *Cache
	readQueue        []*QueueItem
	writeQueue       []*QueueItem
	lastMs           int
}

func createCache(diskMemory int, isRam bool, destCache *Cache) *Cache {
	cache := new(Cache)
	cache.functionMap = make(map[string]*FunctionInCache)
	cache.diskMemory = diskMemory
	cache.isRam = isRam
	cache.destCache = destCache
	cache.occupied = 0
	cache.lastMs = 0
	return cache
}

func updateReadQueue(diskCache *Cache, ms int) {
	i := 0
	for ; i < len(diskCache.readQueue); i++ {
		item := diskCache.readQueue[i]
		if item.transferEnd > ms {
			break
		}
		diskCache.diskMemory += item.memory
	}
	diskCache.readQueue = diskCache.readQueue[i:]
}

func updateWriteQueue(diskCache *Cache, ms int, node *Node) {
	i := 0
	for ; i < len(diskCache.writeQueue); i++ {
		item := diskCache.writeQueue[i]
		if item.transferEnd > ms {
			break
		}
		insertCacheItem(diskCache, item.function, item.memory, item.transferEnd)
		node.ramMemory += item.memory
	}
	diskCache.writeQueue = diskCache.writeQueue[i:]
}

func addToReadQueue(diskCache *Cache, function string, memory int, ms int) int {
	retrieveCache(diskCache, function)
	transfer := int(float32(memory) / props.readBandwidth)
	queueItem := new(QueueItem)
	queueItem.function = function
	queueItem.memory = memory
	if len(diskCache.readQueue) == 0 {
		queueItem.transferEnd = ms + transfer
	} else {
		queueItem.transferEnd = transfer + diskCache.readQueue[len(diskCache.readQueue)-1].transferEnd
	}

	if queueItem.transferEnd-ms >= 250*0.8 {
		queueItem = nil
		return -1
	}
	diskCache.readQueue = append(diskCache.readQueue, queueItem)
	return queueItem.transferEnd - ms
}

func addToWriteQueue(diskCache *Cache, function string, memory int, ms int) {
	transfer := int(float32(memory) / props.writeBandwidth)
	queueItem := new(QueueItem)
	queueItem.function = function
	queueItem.memory = memory
	if len(diskCache.writeQueue) == 0 {
		queueItem.transferEnd = ms + transfer
	} else {
		queueItem.transferEnd = transfer + diskCache.writeQueue[len(diskCache.writeQueue)-1].transferEnd
	}
	diskCache.writeQueue = append(diskCache.writeQueue, queueItem)
}

func freeCache(cache *Cache, memory int, ms int) int {
	i := 0
	freedMem := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if freedMem > memory {
			break
		}
		freedMem += cache.functionMap[cache.orderedFunctions[i].function].memory
		cache.functionMap[cache.orderedFunctions[i].function].copies--

		if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
			delete(cache.functionMap, cache.orderedFunctions[i].function)
		}

		cache.orderedFunctions[i] = nil
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
	if !cache.isRam {
		cache.diskMemory += freedMem
	} else {
		cache.occupied -= freedMem
	}
	return freedMem
}

func insertCacheItem(cache *Cache, name string, memory int, start int) {

	if !cache.isRam {
		if cache.diskMemory < memory {
			if freeCache(cache, memory-cache.diskMemory, start) < memory-cache.diskMemory {
				return
			}
		}
	}

	cache.occupied += memory

	//Insert in map
	_, exists := cache.functionMap[name]

	if exists {
		cache.functionMap[name].copies++

	} else {
		newFunction := new(FunctionInCache)
		newFunction.name = name
		newFunction.memory = memory
		newFunction.copies = 1
		cache.functionMap[name] = newFunction
	}

	//Insert in ordered list
	newItem := new(CacheItem)
	newItem.end = start + minToMs(props.keepAliveWindow)
	newItem.function = name
	cache.orderedFunctions = append(cache.orderedFunctions, newItem)

	if !cache.isRam {
		cache.diskMemory -= memory
	}
}

func freeBuffer(cache *Cache, memory int) int {
	i := 0
	freedMem := 0
	for ; i < len(cache.writeQueue); i++ {
		if freedMem > memory {
			break
		}
		freedMem += cache.writeQueue[i].memory
		cache.writeQueue[i] = nil
	}
	cache.writeQueue = cache.writeQueue[i:]

	return freedMem
}

func updateCache(node *Node, cache *Cache, ms int) {

	if cache.isRam {

		for cache.lastMs < ms-100 {

			if float32(cache.occupied)/float32(node.MAX_MEMORY) > 0.5 {
				i := 0
				tempMem := cache.occupied
				for ; i < len(cache.orderedFunctions); i++ {
					if float32(tempMem)/float32(node.MAX_MEMORY) <= 0.5 {
						break
					}
					cache.functionMap[cache.orderedFunctions[i].function].copies--
					item := cache.functionMap[cache.orderedFunctions[i].function]
					addToWriteQueue(cache.destCache, item.name, item.memory, ms)
					node.ramCache.occupied -= item.memory
					tempMem -= cache.functionMap[cache.orderedFunctions[i].function].memory
					if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
						delete(cache.functionMap, cache.orderedFunctions[i].function)
					}
					cache.orderedFunctions[i] = nil
				}
				cache.orderedFunctions = cache.orderedFunctions[i:]
			}
			cache.lastMs = cache.lastMs + 100
		}

	}

	if !cache.isRam {
		updateReadQueue(cache, ms)
		updateWriteQueue(cache, ms, node)
	}
}

func retrieveCache(cache *Cache, name string) {
	cache.functionMap[name].copies--
	cache.occupied -= cache.functionMap[name].memory
	if cache.functionMap[name].copies == 0 {
		delete(cache.functionMap, name)
	}
	i := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if cache.orderedFunctions[i].function == name {
			break
		}
	}
	cache.orderedFunctions[i] = nil
	cache.orderedFunctions = append(cache.orderedFunctions[:i], cache.orderedFunctions[i+1:]...)
}

func searchCache(cache *Cache, name string) bool {
	_, exists := cache.functionMap[name]
	if exists {
		return true
	}
	return false
}
