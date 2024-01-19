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
	occupied         int
	destCache        *DiskCache
	lastMs           int
}
type DiskCache struct {
	functionMap      map[string]*FunctionInCache
	orderedFunctions []*CacheItem
	diskMemory       int
	occupied         int
	readQueue        []*QueueItem
	writeQueue       []*QueueItem
	lastMs           int
}

func createCache(destCache *DiskCache) *Cache {
	cache := new(Cache)
	cache.functionMap = make(map[string]*FunctionInCache)
	cache.destCache = destCache
	cache.occupied = 0
	cache.lastMs = 0
	return cache
}

func createDisk(diskMemory int) *DiskCache {
	diskCache := new(DiskCache)
	diskCache.functionMap = make(map[string]*FunctionInCache)
	diskCache.diskMemory = diskMemory
	diskCache.occupied = 0
	diskCache.lastMs = 0
	return diskCache
}

func updateReadQueue(diskCache *DiskCache, ms int) {
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

func updateWriteQueue(diskCache *DiskCache, ms int, node *Node) {
	i := 0
	for ; i < len(diskCache.writeQueue); i++ {
		item := diskCache.writeQueue[i]
		if item.transferEnd > ms {
			break
		}
		insertDiskItem(diskCache, item.function, item.memory, item.transferEnd)
		node.ramMemory += item.memory
	}
	diskCache.writeQueue = diskCache.writeQueue[i:]
}

func addToReadQueue(diskCache *DiskCache, function string, memory int, ms int) int {
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

func addToWriteQueue(diskCache *DiskCache, function string, memory int, ms int) {

	if searchDisk(diskCache, function) == true {
		return
	}

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

func freeCache(cache *Cache, memory int) int {
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
	cache.occupied -= freedMem
	return freedMem
}

func freeDiskCache(cache *DiskCache, memory int) int {
	i := 0
	freedMem := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if freedMem > memory {
			break
		}
		freedMem += cache.functionMap[cache.orderedFunctions[i].function].memory
		delete(cache.functionMap, cache.orderedFunctions[i].function)
		cache.orderedFunctions[i] = nil
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
	cache.diskMemory += freedMem
	return freedMem
}

func insertCacheItem(cache *Cache, name string, memory int, start int) {

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
}

func insertDiskItem(diskCache *DiskCache, name string, memory int, start int) {

	if diskCache.diskMemory < memory {
		if freeDiskCache(diskCache, memory-diskCache.diskMemory) < memory-diskCache.diskMemory {
			return
		}
	}

	diskCache.occupied += memory

	//Insert in map
	_, exists := diskCache.functionMap[name]

	if exists {
		return

	} else {
		newFunction := new(FunctionInCache)
		newFunction.name = name
		newFunction.memory = memory
		newFunction.copies = 1
		diskCache.functionMap[name] = newFunction
	}

	//Insert in ordered list
	newItem := new(CacheItem)
	newItem.end = start + minToMs(props.keepAliveWindow)
	newItem.function = name
	diskCache.orderedFunctions = append(diskCache.orderedFunctions, newItem)

	diskCache.diskMemory -= memory
}

func freeBuffer(diskCache *DiskCache, memory int) int {
	i := 0
	freedMem := 0
	for ; i < len(diskCache.writeQueue); i++ {
		if freedMem > memory {
			break
		}
		freedMem += diskCache.writeQueue[i].memory
		diskCache.writeQueue[i] = nil
	}
	diskCache.writeQueue = diskCache.writeQueue[i:]

	return freedMem
}

func updateCache(node *Node, cache *Cache, ms int) {

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

func updateDisk(node *Node, diskCache *DiskCache, ms int) {
	updateReadQueue(diskCache, ms)
	updateWriteQueue(diskCache, ms, node)
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

func searchDisk(diskCache *DiskCache, name string) bool {
	_, exists := diskCache.functionMap[name]
	if exists {
		return true
	}
	return false
}
