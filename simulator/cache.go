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

type Cache struct {
	functionMap      map[string]*FunctionInCache
	orderedFunctions []*CacheItem
	memory           int
	isRam            bool
	destCache        *Cache
}

func createCache(memory int, isRam bool, destCache *Cache) *Cache {
	cache := new(Cache)
	cache.functionMap = make(map[string]*FunctionInCache)
	cache.memory = memory
	cache.isRam = isRam
	cache.destCache = destCache
	return cache
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

		if cache.isRam {
			item := cache.functionMap[cache.orderedFunctions[i].function]
			insertCacheItem(cache.destCache, item.name, item.memory, ms)
		}

		if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
			delete(cache.functionMap, cache.orderedFunctions[i].function)
		}

		cache.orderedFunctions[i] = nil
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
	cache.memory += freedMem
	return freedMem
}

func insertCacheItem(cache *Cache, name string, memory int, start int) {

	if cache.memory < memory {
		if freeCache(cache, memory-cache.memory, start) < memory-cache.memory {
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

func updateCache(cache *Cache, ms int) {
	i := 0
	for ; i < len(cache.orderedFunctions); i++ {
		if cache.orderedFunctions[i].end > ms {
			break
		} else {
			cache.functionMap[cache.orderedFunctions[i].function].copies--
			cache.memory += cache.functionMap[cache.orderedFunctions[i].function].memory

			if cache.isRam {
				item := cache.functionMap[cache.orderedFunctions[i].function]
				insertCacheItem(cache.destCache, item.name, item.memory, ms)
			}

			if cache.functionMap[cache.orderedFunctions[i].function].copies == 0 {
				delete(cache.functionMap, cache.orderedFunctions[i].function)
			}
			cache.orderedFunctions[i] = nil
		}
	}
	cache.orderedFunctions = cache.orderedFunctions[i:]
}

func retrieveCache(cache *Cache, name string) {
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

func searchCache(cache *Cache, name string) bool {
	_, exists := cache.functionMap[name]
	if exists {
		return true
	}
	return false
}
