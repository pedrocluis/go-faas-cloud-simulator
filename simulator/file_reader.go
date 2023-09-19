package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

// Atoi function with error checking
func atoi(s string) int {
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func readFile(filename string) []Invocation {

	var invocations []Invocation

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
		var element Invocation
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
		element.hashOwner = rec[0]
		element.hashFunction = rec[1]
		element.memory = atoi(rec[2])
		element.duration = atoi(rec[3])
		element.timestamp = atoi(rec[4])
		invocations = append(invocations, element)
	}

	return invocations

}

func readLines(csvReader *csv.Reader, isFirstLine bool) []Invocation {
	//Read csv file line by line
	var invocations []Invocation

	for i := 0; i < MAX_INVOCATIONS; i++ {
		var element Invocation
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if isFirstLine {
			isFirstLine = false
			i--
			continue
		}
		//Create the object
		element.hashOwner = rec[0]
		element.hashFunction = rec[1]
		element.memory = atoi(rec[2])
		element.duration = atoi(rec[3])
		element.timestamp = atoi(rec[4])
		invocations = append(invocations, element)
	}
	return invocations
}
