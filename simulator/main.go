package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const N_NODES = 80
const RUN_MEMORY = 32000
const RAM_MEMORY = 10000
const DISK_MEMORY = 250000
const N_THREADS = 4
const INPUT_FILE = "dataset/trace_d01_1_30.txt"
const KEEP_ALIVE_WINDOW = 5
const STAT_FILE = "data.csv"
const COLD_LATENCY = 250
const READ_BANDWIDTH = 10
const WRITE_BANDWIDTH = 10
const MAX_INVOCATIONS = 1000000

var props *Properties

func alloc_loop(nodeList *[]Node, stats *Statistics, lock *sync.Mutex, idx *int, invocations []Invocation) {

	//Place invocations one by one
	for {
		lock.Lock()
		*idx++
		if *idx >= len(invocations) {
			lock.Unlock()
			break
		}
		i := *idx
		invocation := Invocation{
			hashOwner:    invocations[i].hashOwner,
			hashFunction: invocations[i].hashFunction,
			memory:       invocations[i].memory,
			duration:     invocations[i].duration,
			timestamp:    invocations[i].timestamp,
		}
		chosenNode := findNode(nodeList, invocation.timestamp, stats, invocation.hashFunction)
		stats.invocations[chosenNode]++
		lock.Unlock()
		stats.statsLock.Lock()
		stats.invocationsSecond++
		stats.statsLock.Unlock()
		allocateInvocation(&(*nodeList)[chosenNode], invocation, stats)
	}
}

func threadFunc(wg *sync.WaitGroup, nodeList *[]Node, stats *Statistics, lock *sync.Mutex, globalIndex *int, invocations []Invocation) {

	defer wg.Done()

	alloc_loop(nodeList, stats, lock, globalIndex, invocations)

}

func main() {

	//Read command line arguments
	props = getProperties()

	//Measure the execution time
	timeStart := time.Now()

	//Initialize statistics struct
	stats := new(Statistics)
	createStatistics(stats, props.statFile)

	// List of nodes
	listNodes := make([]Node, props.nNodes)

	// Declare mutex and wait group
	invocationsLock := new(sync.Mutex)
	var wg sync.WaitGroup

	//Create the number of nodes specified
	for num := 0; num < props.nNodes; num++ {
		listNodes[num] = createNode(num, props.runMemory, props.ramMemory)
	}

	f, err := os.Open(props.inputFile)
	if err != nil {
		log.Fatal(err)
	}

	//When the function ends close the file
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	isFirstLine := true

	counter := 0

	for {
		//Add all the threads to the wait group
		wg.Add(props.nThreads)
		fmt.Printf("Invocations: %d\n", counter*MAX_INVOCATIONS)
		invocations := readLines(csvReader, isFirstLine)
		isFirstLine = false

		globalIndex := new(int)
		*globalIndex = -1

		for n := 0; n < props.nThreads; n++ {
			go threadFunc(&wg, &listNodes, stats, invocationsLock, globalIndex, invocations)
		}

		//Wait for the threads to finish
		wg.Wait()

		if len(invocations) < MAX_INVOCATIONS {
			fmt.Println(len(invocations))
			break
		}
		invocations = nil
		counter++
	}

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
