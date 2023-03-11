package main

import (
	"fmt"
)

func addDurations(functionInvocations []functionInvocationCount, durations []functionExecutionDuration) []functionInvocationCount {
	for i := range functionInvocations {
		functionInvocations[i].duration = -1
		for j := range durations {
			if functionInvocations[i].function == durations[j].function {
				functionInvocations[i].duration = durations[j].average / 1000 / 60 // Convert duration from ms to minute
				break
			}
		}
	}
	return functionInvocations
}

func addMemories(functionInvocations []functionInvocationCount, memoryUsages []appMemory) []functionInvocationCount {
	for i := range functionInvocations {
		functionInvocations[i].memory = -1
		for j := range memoryUsages {
			if functionInvocations[i].app == memoryUsages[j].app {
				functionInvocations[i].memory = memoryUsages[j].average
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

		if i%5000 == 0 {
			fmt.Printf("%d\n", i)
		}

		if listInvocations[i].memory == -1 || listInvocations[i].duration == -1 {
			continue
		}
		/*memory := -1
		duration := -1
		for j := range listMemory {
			if listInvocations[i].app == listMemory[j].app {
				memory = listMemory[j].average
				break
			}
		}
		if memory == -1 {
			continue
		}
		for k := range functionDuration {
			if listInvocations[i].function == functionDuration[k].function {
				duration = functionDuration[k].average
				break
			}
		}
		if duration == -1 {
			continue
		}
		// Convert duration from ms to minute
		duration = duration / 1000 / 60
		*/
		for l := range listInvocations[i].perMinute {
			if listInvocations[i].perMinute[l] != 0 {
				invocations++
				allocateMemory(n, listInvocations[i].app, l, listInvocations[i].memory, listInvocations[i].duration, &coldStarts)
			}
		}
	}

	fmt.Printf("Invocations: %d\n", invocations)
	fmt.Printf("Cold Starts: %d\n", coldStarts)
}
