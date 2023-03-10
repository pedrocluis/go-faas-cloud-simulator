package main

import (
	"fmt"
)

func main() {
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
	fmt.Println(listInvocations[0])
	listMemory := readAppMemoryCsvFile("dataset/app_memory_percentiles.anon.d01.csv")
	fmt.Println(listMemory[0])
	durationMemory := readFunctionDurationCsvFile("dataset/function_durations_percentiles.anon.d09.csv")
	fmt.Println(durationMemory[0])

}
