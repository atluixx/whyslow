package analyzers

import (
	"runtime"
	"sort"

	"github.com/atluixx/whyslow/internal/models"
)

func ProcessUsage(
	oldProcesses, nowProcesses []models.ProcessStats,
	oldCPU, nowCPU models.CPUStats,
) []models.ProcessStats {
	oldByPID := make(map[int]models.ProcessStats)

	for _, process := range oldProcesses {
		oldByPID[process.PID] = process
	}

	user := nowCPU.User - oldCPU.User
	nice := nowCPU.Nice - oldCPU.Nice
	system := nowCPU.System - oldCPU.System
	idle := nowCPU.Idle - oldCPU.Idle
	iowait := nowCPU.Iowait - oldCPU.Iowait
	irq := nowCPU.IRQ - oldCPU.IRQ
	softirq := nowCPU.SoftIRQ - oldCPU.SoftIRQ
	steal := nowCPU.Steal - oldCPU.Steal

	busy := user + nice + system + irq + softirq + steal
	systemDelta := busy + idle + iowait

	if systemDelta == 0 {
		return nil
	}

	result := make([]models.ProcessStats, 0, len(nowProcesses))

	for _, process := range nowProcesses {
		previous, exists := oldByPID[process.PID]
		if !exists {
			continue
		}

		delta := process.CPUTime - previous.CPUTime

		process.CPUUsage =
			float64(delta) /
				float64(systemDelta) *
				float64(runtime.NumCPU()) *
				100

		result = append(result, process)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CPUUsage > result[j].CPUUsage
	})

	return result
}
