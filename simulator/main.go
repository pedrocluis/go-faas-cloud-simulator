package main

import "fmt"

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
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
	listMemory := readAppMemoryCsvFile("dataset/app_memory_percentiles.anon.d01.csv")
	functionDuration := readFunctionDurationCsvFile("dataset/function_durations_percentiles.anon.d09.csv")

	listInvocations = addDurations(listInvocations, functionDuration)
	listInvocations = addMemories(listInvocations, listMemory)

	coldStarts := 0
	invocations := 0

	n := newNode(1000000000000000000)
	//for i := range listInvocations {
	for i := 0; i < 10000; i++ {

		if i%2500 == 0 {
			fmt.Printf("%d\n", i)
		}

		if listInvocations[i].avgMemory == -1 || listInvocations[i].avgDuration == -1 {
			continue
		}

		for l := range listInvocations[i].perMinute {
			if listInvocations[i].perMinute[l] != 0 {
				invocations++
				allocateMemory(n, listInvocations[i].app, l, listInvocations[i].avgMemory, listInvocations[i].avgDuration, &coldStarts)
			}
		}
	}

	fmt.Printf("Keep Alive: %d\n", KEEP_ALIVE_WINDOW)
	fmt.Printf("Invocations: %d\n", invocations)
	//fmt.Printf("Cold Starts: %d\n", coldStarts)
}
