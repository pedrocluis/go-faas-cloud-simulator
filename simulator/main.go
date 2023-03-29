package main

import (
	"fmt"
	"sync"
	"time"
)

const N_NODES = 100
const NODE_MEMORY = 64000
const N_THREADS = 8
const UNLOAD_POLICY = "random"
const MAX_DATASET_SIZE = 1000

type Statistics struct {
	invocations    [N_NODES]int
	coldStarts     [N_NODES]int
	failed         [N_NODES]int
	minuteProgress [MINUTES_IN_DAY + 1]int
	minutesLock    sync.Mutex
}

// This function adds the average duration of a function to the invocation count structure
func addDurations(functionInvocations []functionInvocationCount, durations []functionExecutionDuration) []functionInvocationCount {
	for i := range functionInvocations {
		functionInvocations[i].avgDuration = -1
		for j := range durations {
			if functionInvocations[i].function == durations[j].function {
				functionInvocations[i].avgDuration = durations[j].average
				break
			}
		}
	}
	return functionInvocations
}

// This function adds the average memory of a function to the invocation count structure
func addMemories(functionInvocations []functionInvocationCount, memoryUsages []appMemory) []functionInvocationCount {
	for i := range functionInvocations {
		functionInvocations[i].avgMemory = -1
		for j := range memoryUsages {
			if functionInvocations[i].app == memoryUsages[j].app {
				functionInvocations[i].avgMemory = memoryUsages[j].average
				break
			}
		}
	}
	return functionInvocations
}

func allocLoop(listInvocations []functionInvocationCount, nodeList [N_NODES]Node, firstNode, lastNode, start int, end int, stats *Statistics) {

	currentNode := firstNode

	// Look at each minute
	for min := 1; min <= MINUTES_IN_DAY; min++ {

		// Look at the functions for this node
		for i := start; i < end; i++ {

			//If the function doesn't have information about the memory or duration we skip it (don't invoke it)
			if listInvocations[i].avgMemory == -1 || listInvocations[i].avgDuration == -1 {
				continue
			}

			// Allocate memory for this minute for each invocation
			for invocationCount := 0; invocationCount < listInvocations[i].perMinute[min]; invocationCount++ {
				invocation := listInvocations[i]
				allocateMemory(&nodeList[currentNode], invocation.app, min, invocation.avgMemory, invocation.avgDuration, stats)
				stats.invocations[currentNode]++
				currentNode++
				if currentNode == lastNode {
					currentNode = firstNode
				}
			}
		}

		stats.minutesLock.Lock()
		stats.minuteProgress[min]++
		if stats.minuteProgress[min] == N_THREADS {
			fmt.Printf("Minute %d completed!\n", min)
		}
		stats.minutesLock.Unlock()

	}

}

func threadFunc(threadN int, firstNode int, lastNode int, wg *sync.WaitGroup, listInvocations []functionInvocationCount, nodeList [N_NODES]Node, stats *Statistics) {

	defer wg.Done()

	allocLoop(listInvocations, nodeList, firstNode, lastNode, threadN*len(listInvocations)/N_THREADS, (threadN+1)*len(listInvocations)/N_THREADS, stats)

}

func main() {

	//Measure the execution time
	timeStart := time.Now()

	//Initialize statistics struct
	stats := new(Statistics)

	//Read the csv files into structure arrays
	fmt.Println("Reading the invocations per function file")
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
	//Cut the dataset
	if MAX_DATASET_SIZE < len(listInvocations) {
		listInvocations = listInvocations[:MAX_DATASET_SIZE]
	}
	fmt.Println("Reading the app memory file")
	listMemory := readAppMemoryCsvFile("dataset/app_memory_percentiles.anon.d01.csv")
	fmt.Println("Reading the function duration file")
	functionDuration := readFunctionDurationCsvFile("dataset/function_durations_percentiles.anon.d09.csv")

	//Add the durations and memory to the invocation structure, so we have everything in the same array
	fmt.Println("Joining the average durations to each function")
	listInvocations = addDurations(listInvocations, functionDuration)
	fmt.Println("Joining the average memory usage to each function")
	listInvocations = addMemories(listInvocations, listMemory)

	// List of nodes
	var listNodes [N_NODES]Node

	// Declare mutex and wait group
	var wg sync.WaitGroup

	//Add all the threads to the wait group
	wg.Add(N_THREADS)
	fmt.Printf("Size of the Dataset: %d\n", len(listInvocations))
	//Create the number of nodes specified and send them to a thread

	for num := 0; num < N_NODES; num++ {
		listNodes[num] = newNode(NODE_MEMORY, 0)
	}

	for n := 0; n < N_THREADS; n++ {
		go threadFunc(n, n*len(listNodes)/N_THREADS, (n+1)*len(listNodes)/N_THREADS, &wg, listInvocations, listNodes, stats)
	}

	//Wait for the threads to finish
	wg.Wait()
	timeElapsed := time.Since(timeStart)

	invocationsSum := 0
	for i := range stats.invocations {
		invocationsSum += stats.invocations[i]
	}
	failedInvocationsSum := 0
	for i := range stats.failed {
		failedInvocationsSum += stats.failed[i]
	}

	fmt.Printf("The simulation took %s\n", timeElapsed)
	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", invocationsSum)
	fmt.Printf("Failed Invocations: %d\n", failedInvocationsSum)
}
