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
	return ProcessUsageWithCPUCount(oldProcesses, nowProcesses, oldCPU, nowCPU, runtime.NumCPU())
}

// ProcessUsageWithCPUCount calculates per-process CPU percentage, where 100%
// represents one fully busy CPU. cpuCount is injectable for deterministic tests.
func ProcessUsageWithCPUCount(
	oldProcesses, nowProcesses []models.ProcessStats,
	oldCPU, nowCPU models.CPUStats,
	cpuCount int,
) []models.ProcessStats {
	oldByPID := make(map[int]models.ProcessStats)

	for _, process := range oldProcesses {
		oldByPID[process.PID] = process
	}

	activity := CPUActivity(oldCPU, nowCPU)
	if activity.TotalTicks == 0 || cpuCount <= 0 {
		return nil
	}

	result := make([]models.ProcessStats, 0, len(nowProcesses))

	for _, process := range nowProcesses {
		previous, exists := oldByPID[process.PID]
		if !exists {
			continue
		}

		// A PID may have been reused, or a process counter may have reset.
		if process.CPUTime < previous.CPUTime {
			continue
		}
		delta := process.CPUTime - previous.CPUTime

		process.CPUUsage =
			float64(delta) /
				float64(activity.TotalTicks) *
				float64(cpuCount) *
				100

		result = append(result, process)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CPUUsage > result[j].CPUUsage
	})

	return result
}
