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

	oldProcesses, err := collectors.ReadProcesses()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(300 * time.Millisecond)

		memoryNow, err := collectors.ReadMemoryStats()
		if err != nil {
			panic(err)
		}

		loadStats, err := collectors.ReadLoadStats()
		if err != nil {
			panic(err)
		}

		cpuNow, err := collectors.ReadCPUStats()
		if err != nil {
			panic(err)
		}

		newProcesses, err := collectors.ReadProcesses()
		if err != nil {
			panic(err)
		}

		memoryUsage := analyzers.MemoryUsage(memoryNow)
		cpuUsage := analyzers.CPUUsage(cpuOld, cpuNow)
		processesWithUsage := analyzers.ProcessUsage(oldProcesses, newProcesses, cpuOld, cpuNow)

		ui.ClearScreen()

		fmt.Printf("CPU Usage:    %6.2f%%\n", cpuUsage)
		fmt.Printf("Memory Usage: %6.2f%%\n", memoryUsage)
		fmt.Printf(
			"Load Average: %.2f %.2f %.2f\n",
			loadStats.Load1,
			loadStats.Load5,
			loadStats.Load15,
		)

		fmt.Println()
		fmt.Println("Top Processes")
		fmt.Println("-------------------------------------------------------------------")
		fmt.Printf(
			"%-8s %-25s %8s %8s %12s\n",
			"PID",
			"COMMAND",
			"%CPU",
			"THREADS",
			"RSS",
		)
		fmt.Println("-------------------------------------------------------------------")

		limit := min(len(processesWithUsage), 10)

		for _, p := range processesWithUsage[:limit] {
			fmt.Printf(
				"%-8d %-25s %7.2f%% %8d %10.2f MB\n",
				p.PID,
				p.Name,
				p.CPUUsage,
				p.Threads,
				float64(p.RSS)/(1024*1024),
			)
		}

		cpuOld = cpuNow
		oldProcesses = newProcesses
	}
}
