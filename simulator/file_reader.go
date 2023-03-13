package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// Atoi function with error checking
func atoi(s string) int {
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

// Read datasets about the invocation count and create an array of functionInvocationCount structs
func readInvocationCsvFile(filename string) []functionInvocationCount {

	var functionInvocationCounts []functionInvocationCount

	//Open the file
	f, err := os.Open(filename)
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

	//Read csv file line by line
	csvReader := csv.NewReader(f)
	isFirstLine := true
	for {
		var element functionInvocationCount
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if isFirstLine {
			isFirstLine = false
			continue
		}
		//Create the object
		element.owner = rec[0]
		element.app = rec[1]
		element.function = rec[2]
		element.trigger = rec[3]
		for i := 1; i < 1441; i++ {
			element.perMinute[i] = atoi(rec[i+3])
		}
		functionInvocationCounts = append(functionInvocationCounts, element)
	}

	return functionInvocationCounts

}

// Read datasets about the app memory usage and create an array of appMemory structs
func readAppMemoryCsvFile(filename string) []appMemory {
	var appMemoryUsages []appMemory

	//Open the file
	f, err := os.Open(filename)
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

	//Read csv file line by line
	csvReader := csv.NewReader(f)
	isFirstLine := true
	for {
		var element appMemory
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if isFirstLine {
			isFirstLine = false
			continue
		}
		//Create the object
		element.owner = rec[0]
		element.app = rec[1]
		element.count = atoi(rec[2])
		element.average = atoi(rec[3])
		element.percentileAverage1 = atoi(rec[4])
		element.percentileAverage5 = atoi(rec[5])
		element.percentileAverage25 = atoi(rec[6])
		element.percentileAverage50 = atoi(rec[7])
		element.percentileAverage75 = atoi(rec[8])
		element.percentileAverage95 = atoi(rec[9])
		element.percentileAverage99 = atoi(rec[10])
		element.percentileAverage100 = atoi(rec[11])

		appMemoryUsages = append(appMemoryUsages, element)
	}

	return appMemoryUsages
}

// Read datasets about the function duration count and create an array of functionExecutionDuration structs
func readFunctionDurationCsvFile(filename string) []functionExecutionDuration {
	var functionDurations []functionExecutionDuration

	//Open the file
	f, err := os.Open(filename)
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

	//Read csv file line by line
	csvReader := csv.NewReader(f)
	isFirstLine := true
	for {
		var element functionExecutionDuration
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if isFirstLine {
			isFirstLine = false
			continue
		}
		//Create the object
		element.owner = rec[0]
		element.app = rec[1]
		element.function = rec[2]
		element.average = atoi(rec[3])
		element.count = atoi(rec[4])
		element.minimum = atoi(strings.Split(rec[5], ".")[0])
		element.maximum = atoi(strings.Split(rec[6], ".")[0])
		element.percentileAverage0 = atoi(rec[7])
		element.percentileAverage1 = atoi(rec[8])
		element.percentileAverage25 = atoi(rec[9])
		element.percentileAverage50 = atoi(rec[10])
		element.percentileAverage75 = atoi(rec[11])
		element.percentileAverage99 = atoi(rec[12])
		element.percentileAverage100 = atoi(rec[13])

		functionDurations = append(functionDurations, element)
	}

	return functionDurations
}
