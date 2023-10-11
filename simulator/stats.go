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

	latencyCdf [N_NODES][]int
}

func createStatistics(stats *Statistics, stat_file string) {
	stats.statsLock = new(sync.Mutex)
	for i := range stats.latencyCdf {
		stats.latencyCdf[i] = make([]int, 0)
	}
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

func writeStats(stats *Statistics, runMemAvg int, ramMemAvg int, second int, diskMemAvg int) {
	f, err := os.OpenFile(stats.statsFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d\n",
		second,
		stats.invocationsSecond,
		stats.failedInvocationsSecond,
		stats.coldStartsSecond,
		stats.warmStartsSecond,
		runMemAvg,
		ramMemAvg,
		stats.lukeWarmStartsSecond,
		diskMemAvg))

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

func write_latencies(stats *Statistics) {

	statsCdfFile := "latencies_" + stats.statsFile

	f, err := os.OpenFile(statsCdfFile,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)

	for n := range stats.latencyCdf {
		for i := range stats.latencyCdf[n] {
			_, err := f.WriteString(fmt.Sprintf("%d\n", stats.latencyCdf[n][i]))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func write_starts(stats *Statistics, starts string) {

	f, err := os.OpenFile(starts+".csv",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer func(f *os.File) {
		err1 := f.Close()
		if err1 != nil {
			fmt.Println(err1)
		}
	}(f)

	var total float32
	var memory int
	if starts == "lukewarm" {
		total = float32(stats.totalLukeWarmStarts) / float32(stats.totalColdStarts+stats.totalWarmStarts+stats.totalLukeWarmStarts) * 100.0
		memory = props.diskMemory
	} else {
		if starts == "warm" {
			total = float32(stats.totalWarmStarts) / float32(stats.totalColdStarts+stats.totalWarmStarts+stats.totalLukeWarmStarts) * 100
			memory = props.ramMemory
		}
	}

	_, err = f.WriteString(fmt.Sprintf("%d,%.0f\n", memory, total))
	if err != nil {
		fmt.Println(err)
	}

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
	for i := range stats.lukewarmStarts {
		stats.totalLukeWarmStarts += stats.lukewarmStarts[i]
	}

	stats.totalFailed = 0
	for i := range stats.failedInvocations {
		stats.totalFailed += stats.failedInvocations[i]
	}

	write_latencies(stats)
	write_starts(stats, "lukewarm")
	write_starts(stats, "warm")
}
