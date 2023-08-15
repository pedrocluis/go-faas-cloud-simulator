package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Statistics struct {
	invocations         [N_NODES]int
	warmStarts          [N_NODES]int
	coldStarts          [N_NODES]int
	lukewarmStarts      [N_NODES]int
	failedInvocations   [N_NODES]int
	totalInvocations    int
	totalWarmStarts     int
	totalColdStarts     int
	totalLukeWarmStarts int
	totalFailed         int

	avgRunMemorySecond      int
	avgRamMemorySecond      int
	invocationsSecond       int
	coldStartsSecond        int
	warmStartsSecond        int
	lukeWarmStartsSecond    int
	failedInvocationsSecond int
	statsLock               *sync.Mutex
	statsFile               string
	statsMs                 int
}

func createStatistics(stats *Statistics, stat_file string) {
	stats.statsLock = new(sync.Mutex)
	create, err := os.Create(stat_file)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = create.Close()
	if err != nil {
		return
	}
	stats.statsFile = stat_file
}

func writeStats(stats *Statistics, runMemAvg int, ramMemAvg int, second int) {
	f, err := os.OpenFile(stats.statsFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d\n",
		second,
		stats.invocationsSecond,
		stats.failedInvocationsSecond,
		stats.coldStartsSecond,
		stats.warmStartsSecond,
		runMemAvg,
		ramMemAvg,
		stats.lukeWarmStartsSecond))

	if err != nil {
		fmt.Println(err.Error())
	}

	err = f.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	stats.invocationsSecond = 0
	stats.failedInvocationsSecond = 0
	stats.coldStartsSecond = 0
	stats.warmStartsSecond = 0
	stats.lukeWarmStartsSecond = 0
}

func computeStats(stats *Statistics) {
	stats.totalInvocations = 0
	for i := range stats.invocations {
		stats.totalInvocations += stats.invocations[i]
	}

	stats.totalWarmStarts = 0
	for i := range stats.warmStarts {
		stats.totalWarmStarts += stats.warmStarts[i]
	}

	stats.totalColdStarts = 0
	for i := range stats.coldStarts {
		stats.totalColdStarts += stats.coldStarts[i]
	}

	stats.totalLukeWarmStarts = 0
	for i := range stats.coldStarts {
		stats.totalLukeWarmStarts += stats.lukewarmStarts[i]
	}

	stats.totalFailed = 0
	for i := range stats.coldStarts {
		stats.totalFailed += stats.failedInvocations[i]
	}
}
