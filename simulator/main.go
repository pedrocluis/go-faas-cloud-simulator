package main

import (
	"fmt"
	"sync"
	"time"
)

const N_NODES = 8
const NODE_MEMORY = 256000

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

func allocLoop(listInvocations []functionInvocationCount, node Node, start int, end int, invocations *int, lock *sync.Mutex, wg *sync.WaitGroup, coldStarts *int, coldStartLock *sync.Mutex, failedInvocations *int, failLock *sync.Mutex, functionsDone *int, progressLock *sync.Mutex) {

	// When the function is done, remove it from the wait group
	defer wg.Done()

	// Look at each minute
	for min := 1; min <= 10; min++ {

		// Look at the functions for this node
		for i := start; i < end; i++ {

			//If the function doesn't have information about the memory or duration we skip it (don't invoke it)
			if listInvocations[i].avgMemory == -1 || listInvocations[i].avgDuration == -1 {
				progressLock.Lock()
				*functionsDone++
				/*if *functionsDone%500000 == 0 {
					fmt.Printf("%d functions done\n", *functionsDone)
				}*/
				progressLock.Unlock()
				continue
			}

			// Allocate memory for this minute for each invocation
			for invocationCount := 0; invocationCount < listInvocations[i].perMinute[min]; invocationCount++ {
				invocation := listInvocations[i]
				allocateMemory(&node, invocation.app, min, invocation.avgMemory, invocation.avgDuration, coldStarts, coldStartLock, failedInvocations, failLock)
				lock.Lock()
				*invocations++
				lock.Unlock()
			}
			progressLock.Lock()
			*functionsDone++
			/*if *functionsDone%500 == 0 {
				fmt.Printf("%d functions done\n", *functionsDone)
			}*/
			progressLock.Unlock()
		}

	}

}

func main() {

	//Cold start counter
	coldStarts := 0
	var coldStartLock sync.Mutex

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

	//Initialize the invocations counter
	invocations := 0

	// Declare mutex and wait group
	var lock sync.Mutex
	var wg sync.WaitGroup

	//Add all the threads to the wait group
	wg.Add(N_NODES)
	fmt.Printf("Size of the Dataset: %d\n", len(listInvocations))
	//Create the number of nodes specified and send them to a thread

	failedInvocations := 0
	var failLock sync.Mutex

	functionsDone := 0
	var progressLock sync.Mutex

	for n := 0; n < N_NODES; n++ {
		listNodes[n] = newNode(NODE_MEMORY)
		//fmt.Printf("Node: %d | Start: %d | End: %d\n", n, n*len(listInvocations)/N_NODES, (n+1)*len(listInvocations)/N_NODES)
		go allocLoop(listInvocations, listNodes[n], n*len(listInvocations)/N_NODES, (n+1)*len(listInvocations)/N_NODES, &invocations, &lock, &wg, &coldStarts, &coldStartLock, &failedInvocations, &failLock, &functionsDone, &progressLock)
	}

	//Wait for the threads to finish
	wg.Wait()
	timeElapsed := time.Since(timeStart)
	fmt.Printf("The simulation took %s\n", timeElapsed)
	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", invocations)
	fmt.Printf("Cold Starts: %d\n", coldStarts)
	fmt.Printf("Failed Invocations %d\n", failedInvocations)
}
