package main

import "fmt"

//This function adds the average duration of a function to the invocation count structure
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

//This function adds the average memory of a function to the invocation count structure
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

func main() {

	//Read the csv files into structure arrays
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
	listMemory := readAppMemoryCsvFile("dataset/app_memory_percentiles.anon.d01.csv")
	functionDuration := readFunctionDurationCsvFile("dataset/function_durations_percentiles.anon.d09.csv")

	//Add the durations and memory to the invocation structure, so we have everything in the same array
	listInvocations = addDurations(listInvocations, functionDuration)
	listInvocations = addMemories(listInvocations, listMemory)

	//Initialize the invocation counters
	invocations := 0

	//Creates one node
	//TODO: Support multiple nodes (maybe already works, have to test) (one thread per node might work)
	n := newNode(1000000000000000000)
	//for i := range listInvocations {
	for i := 0; i < 10000; i++ {

		//Print the progress
		if i%2500 == 0 {
			fmt.Printf("%d\n", i)
		}

		//If the function doesn't have information about the memory or duration we skip it (dont invoke it)
		if listInvocations[i].avgMemory == -1 || listInvocations[i].avgDuration == -1 {
			continue
		}

		//Allocate the memory for each minute
		for l := range listInvocations[i].perMinute {
			if listInvocations[i].perMinute[l] != 0 {
				invocations++
				allocateMemory(&n, listInvocations[i].app, l, listInvocations[i].avgMemory, listInvocations[i].avgDuration)
			}
		}
	}

	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", invocations)
	fmt.Printf("Cold Starts: %d\n", countColdStarts(n))
}
