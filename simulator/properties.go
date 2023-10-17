package main

import "flag"

type Properties struct {
	nNodes          int
	ramMemory       int
	diskMemory      int
	nThreads        int
	inputFile       string
	keepAliveWindow int
	statFile        string
	coldLatency     int
	readBandwidth   float32
	writeBandwidth  float32
}

func getProperties() *Properties {
	nNodesPtr := flag.Int("nodes", -1, "Number of nodes")
	ramPtr := flag.Int("ram", -1, "RAM capacity")
	diskPtr := flag.Int("disk", -1, "Disk cache capacity")
	threadsPtr := flag.Int("threads", -1, "Number of threads")
	inputPtr := flag.String("input", "", "Input file")
	keepAlivePtr := flag.Int("keep_alive", -1, "Keep alive window")
	statsPtr := flag.String("stats", "", "Stats output file")
	coldLatencyPtr := flag.Int("cold_lat", -1, "Cold start latency")
	readBandwidthPtr := flag.Float64("read_speed", -1, "Disk read bandwidth")
	writeBandwidthPtr := flag.Float64("write_speed", -1, "Disk write bandwidth")

	flag.Parse()

	props = new(Properties)

	if *nNodesPtr < 0 {
		props.nNodes = N_NODES
	} else {
		props.nNodes = *nNodesPtr
	}

	if *ramPtr < 0 {
		props.ramMemory = RAM
	} else {
		props.ramMemory = *ramPtr
	}

	if *diskPtr < 0 {
		props.diskMemory = DISK_MEMORY
	} else {
		props.diskMemory = *diskPtr
	}

	if *threadsPtr < 0 {
		props.nThreads = N_THREADS
	} else {
		props.nThreads = *threadsPtr
	}

	if *inputPtr == "" {
		props.inputFile = INPUT_FILE
	} else {
		props.inputFile = *inputPtr
	}

	if *keepAlivePtr < 0 {
		props.keepAliveWindow = KEEP_ALIVE_WINDOW
	} else {
		props.keepAliveWindow = *keepAlivePtr
	}

	if *statsPtr == "" {
		props.statFile = STAT_FILE
	} else {
		props.statFile = *statsPtr
	}

	if *coldLatencyPtr < 0 {
		props.coldLatency = COLD_LATENCY
	} else {
		props.coldLatency = *coldLatencyPtr
	}

	if *readBandwidthPtr < 0 {
		props.readBandwidth = READ_BANDWIDTH
	} else {
		props.readBandwidth = float32(*readBandwidthPtr)
	}

	if *writeBandwidthPtr < 0 {
		props.writeBandwidth = WRITE_BANDWIDTH
	} else {
		props.writeBandwidth = float32(*writeBandwidthPtr)
	}

	return props

}
