package main

import (
	"fmt"
	"strings"
)

func main() {
	listInvocations := readInvocationCsvFile("dataset/invocations_per_function_md.anon.d01.csv")
	fmt.Println(listInvocations[0])
	listMemory := readAppMemoryCsvFile("dataset/app_memory_percentiles.anon.d01.csv")
	fmt.Println(listMemory[0])
	fmt.Println(atoi(strings.Split("107.0", ".")[0]))

}
