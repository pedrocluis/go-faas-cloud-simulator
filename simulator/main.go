package main

import (
	"fmt"
	"sync"
	"time"
)

const N_NODES = 4

var functionsDone int
var progressLock sync.Mutex

// This function adds the average duration of a function to the invocation count structure
func addDurations(functionInvocations []functionInvocationCount, durations []functionExecutionDuration) []functionInvocationCount {
	for i := range functionInvocations {
		functionInvocations[i].avgDuration = -1
		for j := range durations {
			if functionInvocations[i].function == durations[j].function {
				functionInvocations[i].avgDuration = durations[j].average / 1000 / 60 // Convert duration from ms to minute
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

func allocLoop(listInvocations []functionInvocationCount, node Node, start int, end int, invocations *int, lock *sync.Mutex, wg *sync.WaitGroup) {

	// When the function is done, remove it from the wait group
	defer wg.Done()

	for i := start; i < end; i++ {

		//If the function doesn't have information about the memory or duration we skip it (don't invoke it)
		if listInvocations[i].avgMemory == -1 || listInvocations[i].avgDuration == -1 {
			continue
		}

		//Allocate the memory for each minute
		for l := range listInvocations[i].perMinute {
			if listInvocations[i].perMinute[l] != 0 {
				//Lock
				lock.Lock()
				*invocations++
				//Unlock
				lock.Unlock()
				allocateMemory(&node, listInvocations[i].app, l, listInvocations[i].avgMemory, listInvocations[i].avgDuration)
			}
		}

		// This is here only to know the progress
		progressLock.Lock()
		functionsDone++
		if functionsDone%5000 == 0 {
			fmt.Printf("%d\n", functionsDone)
		}
		progressLock.Unlock()
	}
}

func main() {

	//Measure the execution time
	timeStart := time.Now()

	//Read the csv files into structure arrays
	fmt.Println("Reading the invocations per function file")
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
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

	//Initialize the invocation counters
	invocations := 0
	functionsDone = 0

	// Declare mutex and wait group
	var lock sync.Mutex
	var wg sync.WaitGroup

	//Add all the threads to the wait group
	wg.Add(N_NODES)
	fmt.Printf("Size of the Dataset: %d\n", len(listInvocations))
	//Create the number of nodes specified and send them to a thread
	for n := 0; n < N_NODES; n++ {
		listNodes[n] = newNode(1000000000000000000)
		fmt.Printf("Node: %d | Start: %d | End: %d\n", n, n*len(listInvocations)/N_NODES, (n+1)*len(listInvocations)/N_NODES)
		go allocLoop(listInvocations, listNodes[n], n*len(listInvocations)/N_NODES, (n+1)*len(listInvocations)/N_NODES, &invocations, &lock, &wg)
	}

	//Wait for the threads to finish
	wg.Wait()
	timeElapsed := time.Since(timeStart)
	fmt.Printf("The simulation took %s", timeElapsed)
	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", invocations)
	//fmt.Printf("Cold Starts: %d\n", countColdStarts(n))
}
