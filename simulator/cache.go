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

type RAMCache struct {
	functionMap      map[string]*FunctionInCache
	orderedFunctions []*CacheItem
	memory           int
}

func createRAMCache(memory int) *RAMCache {
	cache := new(RAMCache)
	cache.functionMap = make(map[string]*FunctionInCache)
	cache.memory = memory
	return cache
}

func freeCache(cache *RAMCache, memory int) int {
	i := 0
	freedMem := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if freedMem > memory {
			break
		}
		_, exists := cache.functionMap[cache.orderedFunctions[i].function]
		if !exists {
			println("OOPSIE")
		}
		freedMem += cache.functionMap[cache.orderedFunctions[i].function].memory
		cache.functionMap[cache.orderedFunctions[i].function].copies--
		if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
			delete(cache.functionMap, cache.orderedFunctions[i].function)
		}
		cache.orderedFunctions[i] = nil
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
	cache.memory += freedMem
	return freedMem
}

func insertRAMItem(cache *RAMCache, name string, memory int, start int) {

	if cache.memory < memory {
		if freeCache(cache, memory-cache.memory) < memory-cache.memory {
			return
		}
	}

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
	newItem.end = start + minToMs(KEEP_ALIVE_WINDOW)
	newItem.function = name
	cache.orderedFunctions = append(cache.orderedFunctions, newItem)

	cache.memory -= memory
}

func updateRAMCache(cache *RAMCache, ms int) {
	i := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if cache.orderedFunctions[i].end > ms {
			break
		} else {
			cache.functionMap[cache.orderedFunctions[i].function].copies--
			cache.memory += cache.functionMap[cache.orderedFunctions[i].function].memory
			if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
				delete(cache.functionMap, cache.orderedFunctions[i].function)
			}
			cache.orderedFunctions[i] = nil
		}
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
}

func retrieveRAMCache(cache *RAMCache, name string) {
	cache.functionMap[name].copies--
	cache.memory += cache.functionMap[name].memory
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

func searchRAMCache(cache *RAMCache, name string) bool {
	_, exists := cache.functionMap[name]
	if exists {
		return true
	}
	return false
}
