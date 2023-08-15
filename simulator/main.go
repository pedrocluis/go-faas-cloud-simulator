package main

import (
	"fmt"
	"sync"
	"time"
)

const N_NODES = 80
const RUN_MEMORY = 48000
const RAM_MEMORY = 8000
const DISK_MEMORY = 250000
const N_THREADS = 8
const INPUT_FILE = "dataset/trace_d01_1_30.txt"
const KEEP_ALIVE_WINDOW = 1
const STAT_FILE = "stats/data.csv"

type Invocation struct {
	hashOwner    string
	hashFunction string
	memory       int
	duration     int
	timestamp    int
}

func alloc_loop(invocations []Invocation, nodeList *[N_NODES]Node, stats *Statistics, lock *sync.Mutex, idx *int) {

	//Place invocations one by one
	for {
		lock.Lock()
		*idx++
		if *idx >= len(invocations) {
			lock.Unlock()
			break
		}
		invocation := Invocation{
			hashOwner:    invocations[*idx].hashOwner,
			hashFunction: invocations[*idx].hashFunction,
			memory:       invocations[*idx].memory,
			duration:     invocations[*idx].duration,
			timestamp:    invocations[*idx].timestamp,
		}
		chosenNode := findNode(nodeList, invocation.timestamp, stats)
		stats.invocations[chosenNode]++
		lock.Unlock()
		stats.statsLock.Lock()
		stats.invocationsSecond++
		stats.statsLock.Unlock()
		allocateInvocation(&nodeList[chosenNode], invocation, stats)
	}
}

func threadFunc(wg *sync.WaitGroup, invocations []Invocation, nodeList *[N_NODES]Node, stats *Statistics, lock *sync.Mutex, globalIndex *int) {

	defer wg.Done()

	alloc_loop(invocations, nodeList, stats, lock, globalIndex)

}

func main() {

	//Measure the execution time
	timeStart := time.Now()

	//Initialize statistics struct
	stats := new(Statistics)
	createStatistics(stats, STAT_FILE)

	// List of nodes
	var listNodes [N_NODES]Node

	// Declare mutex and wait group
	invocationsLock := new(sync.Mutex)
	var wg sync.WaitGroup

	//Add all the threads to the wait group
	wg.Add(N_THREADS)

	//Create the number of nodes specified
	for num := 0; num < N_NODES; num++ {
		listNodes[num] = createNode(num, RUN_MEMORY, RAM_MEMORY)
	}

	//Read input trace
	invocations := readFile(INPUT_FILE)
	println(len(invocations))

	globalIndex := new(int)
	*globalIndex = -1

	for n := 0; n < N_THREADS; n++ {
		go threadFunc(&wg, invocations, &listNodes, stats, invocationsLock, globalIndex)
	}

	//Wait for the threads to finish
	wg.Wait()
	timeElapsed := time.Since(timeStart)

	computeStats(stats)

	fmt.Printf("The simulation took %s\n", timeElapsed)
	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", stats.totalInvocations)
	fmt.Printf("Warm Starts: %d\n", stats.totalWarmStarts)
	fmt.Printf("Cold Starts: %d\n", stats.totalColdStarts)
	fmt.Printf("Lukewarm Starts: %d\n", stats.totalLukeWarmStarts)
	fmt.Printf("Failed Invocations: %d\n", stats.totalFailed)
}
