package main

import (
	"fmt"
	"time"

	"github.com/atluixx/whyslow/internal/analyzers"
	"github.com/atluixx/whyslow/internal/collectors"
	"github.com/atluixx/whyslow/internal/ui"
)

func main() {
	old, err := collectors.ReadCPUStats()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(300 * time.Millisecond)

		now, err := collectors.ReadCPUStats()
		if err != nil {
			panic(err)
		}

		usage := analyzers.CPUUsage(old, now)

		ui.ClearScreen()
		fmt.Printf("CPU Usage: %.2f%%\n", usage)

		old = now
	}
}
