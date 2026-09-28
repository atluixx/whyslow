package collectors

import (
	"time"

	"github.com/atluixx/whyslow/internal/models"
)

// CollectSnapshot gathers all required system telemetry for a single sample.
// Optional temperature data never prevents a snapshot from being collected.
func CollectSnapshot() (models.Snapshot, error) {
	cpu, err := ReadCPUStats()
	if err != nil {
		return models.Snapshot{}, err
	}
	memory, err := ReadMemoryStats()
	if err != nil {
		return models.Snapshot{}, err
	}
	load, err := ReadLoadStats()
	if err != nil {
		return models.Snapshot{}, err
	}
	processes, err := ReadProcesses()
	if err != nil {
		return models.Snapshot{}, err
	}
	disk, err := ReadDiskStats()
	if err != nil {
		return models.Snapshot{}, err
	}
	swap, err := ReadSwapActivity()
	if err != nil {
		return models.Snapshot{}, err
	}
	return models.Snapshot{Timestamp: time.Now(), CPU: cpu, Memory: memory, Load: load, Processes: processes, Disk: disk, SwapActivity: swap, Temperatures: ReadTemperatures()}, nil
}
