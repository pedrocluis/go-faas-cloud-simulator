package main

import (
	"fmt"
	"sync"
	"time"
)

const N_NODES = 80
const RUN_MEMORY = 32000
const RAM_MEMORY = 16000
const N_THREADS = 8
const INPUT_FILE = "dataset/trace_d01_1_30.txt"
const KEEP_ALIVE_WINDOW = 10

type Statistics struct {
	invocations       [N_NODES]int
	warmStarts        [N_NODES]int
	coldStarts        [N_NODES]int
	failedInvocations [N_NODES]int
	totalInvocations  int
	totalWarmStarts   int
	totalColdStarts   int
	totalFailed       int
}

type Invocation struct {
	hashOwner    string
	hashFunction string
	memory       int
	duration     int
	timestamp    int
}

type ThreadRange struct {
	startNode int
	endNode   int
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
		chosenNode := findNode(nodeList, invocation.timestamp)
		stats.invocations[chosenNode]++
		lock.Unlock()
		allocateInvocation(&nodeList[chosenNode], invocation, stats)
	}
}

func threadFunc(wg *sync.WaitGroup, invocations []Invocation, nodeList *[N_NODES]Node, stats *Statistics, lock *sync.Mutex, globalIndex *int) {

	defer wg.Done()

	alloc_loop(invocations, nodeList, stats, lock, globalIndex)

}

func computeStats(stats *Statistics) {
	stats.totalInvocations = 0
	for i := range stats.invocations {
		stats.totalInvocations += stats.invocations[i]
	}

	stats.totalWarmStarts = 0
	for i := range stats.warmStarts {
		stats.totalWarmStarts += stats.warmStarts[i]
	}

	stats.totalColdStarts = 0
	for i := range stats.coldStarts {
		stats.totalColdStarts += stats.coldStarts[i]
	}

	stats.totalFailed = 0
	for i := range stats.coldStarts {
		stats.totalFailed += stats.failedInvocations[i]
	}
}

func main() {

	//Measure the execution time
	timeStart := time.Now()

	//Initialize statistics struct
	stats := new(Statistics)

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
	fmt.Printf("Failed Invocations: %d\n", stats.totalFailed)
}
