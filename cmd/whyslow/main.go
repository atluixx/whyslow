package main

import (
	"fmt"
	"time"

	"github.com/atluixx/whyslow/internal/analyzers"
	"github.com/atluixx/whyslow/internal/collectors"
	"github.com/atluixx/whyslow/internal/ui"
)

func main() {
	cpuOld, err := collectors.ReadCPUStats()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(300 * time.Millisecond)

		memoryNow, err := collectors.ReadMemoryStats()
		if err != nil {
			panic(err)
		}

		cpuNow, err := collectors.ReadCPUStats()
		if err != nil {
			panic(err)
		}

		memoryUsage := analyzers.MemoryUsage(memoryNow)
		cpuUsage := analyzers.CPUUsage(cpuOld, cpuNow)

		ui.ClearScreen()
		fmt.Printf("Memory Usage: %.2f%%\n", memoryUsage)
		fmt.Printf("CPU Usage: %.2f%%\n", cpuUsage)

		cpuOld = cpuNow
	}
}
