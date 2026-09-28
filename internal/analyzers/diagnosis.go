package analyzers

import (
	"fmt"

	"github.com/atluixx/whyslow/internal/models"
)

// Diagnose turns a telemetry snapshot into concise, actionable observations.
func Diagnose(activity models.CPUActivity, memory models.MemoryStats, load models.LoadStats, cpuCount int, processes []models.ProcessStats) []models.Diagnosis {
	result := make([]models.Diagnosis, 0, 5)
	if activity.Usage >= 85 {
		result = append(result, models.Diagnosis{Level: "critical", Title: "CPU saturation", Details: fmt.Sprintf("CPU is %.0f%% busy.", activity.Usage)})
	}
	if activity.Iowait >= 15 {
		result = append(result, models.Diagnosis{Level: "warning", Title: "Storage pressure", Details: fmt.Sprintf("%.0f%% of CPU time is waiting for I/O.", activity.Iowait)})
	}
	if MemoryUsage(memory) >= 90 {
		result = append(result, models.Diagnosis{Level: "critical", Title: "Memory pressure", Details: fmt.Sprintf("Memory is %.0f%% used; applications may be reclaiming memory.", MemoryUsage(memory))})
	}
	if SwapUsage(memory) >= 10 {
		result = append(result, models.Diagnosis{Level: "warning", Title: "Swap in use", Details: fmt.Sprintf("Swap is %.0f%% used; sustained growth can make the system feel slow.", SwapUsage(memory))})
	}
	if cpuCount > 0 && load.Load1 >= float64(cpuCount) {
		result = append(result, models.Diagnosis{Level: "warning", Title: "Runnable queue is long", Details: fmt.Sprintf("1-minute load %.2f meets or exceeds %d available CPUs.", load.Load1, cpuCount)})
	}
	if len(processes) > 0 && processes[0].CPUUsage >= 25 {
		p := processes[0]
		result = append(result, models.Diagnosis{Level: "info", Title: "Busy process", Details: fmt.Sprintf("%s (PID %d) is using %.0f%% CPU.", p.Name, p.PID, p.CPUUsage)})
	}
	if len(result) == 0 {
		return []models.Diagnosis{{Level: "ok", Title: "No immediate bottleneck", Details: "CPU, memory, I/O wait, and load are within normal thresholds."}}
	}
	return result
}
