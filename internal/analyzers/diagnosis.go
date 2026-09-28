package analyzers

import (
	"fmt"
	"runtime"

	"github.com/atluixx/whyslow/internal/models"
)

type Analysis struct {
	Activity             models.CPUActivity
	MemoryUsage          float64
	SwapUsed             uint64
	SwapUsage            float64
	Disk                 DiskActivity
	SwapPagesPerSec      float64
	Processes            []models.ProcessStats
	Temperatures         []models.Temperature
	MaxTemperature       float64
	TemperatureAvailable bool
	Diagnoses            []models.Diagnosis
	Events               []models.Event
}

// Analyze compares two snapshots, updates bounded history, and explains only
// conditions that have persisted for several samples.
func Analyze(old, now models.Snapshot, history *History, cpuCount int) Analysis {
	if cpuCount <= 0 {
		cpuCount = runtime.NumCPU()
	}
	interval := now.Timestamp.Sub(old.Timestamp)
	activity := CPUActivity(old.CPU, now.CPU)
	processes := ProcessUsageWithCPUCount(old.Processes, now.Processes, old.CPU, now.CPU, cpuCount)
	disk := DiskUsage(old.Disk, now.Disk, interval)
	maxTemperature, hasTemperature := maximumTemperature(now.Temperatures)
	sample := models.Sample{
		Timestamp: now.Timestamp, CPUUsage: activity.Usage, MemoryUsage: MemoryUsage(now.Memory), SwapUsage: SwapUsage(now.Memory), Load1: now.Load.Load1,
		ReadBytesPerSec: disk.ReadBytesPerSec, WriteBytesPerSec: disk.WriteBytesPerSec,
		SwapPagesPerSec: swapPagesPerSecond(old.SwapActivity, now.SwapActivity, interval), MaxTemperature: maxTemperature,
	}
	if len(processes) > 0 {
		sample.TopProcess = processes[0].Name
	}
	history.AddSample(sample)

	analysis := Analysis{Activity: activity, MemoryUsage: sample.MemoryUsage, SwapUsed: SwapUsed(now.Memory), SwapUsage: sample.SwapUsage, Disk: disk, SwapPagesPerSec: sample.SwapPagesPerSec, Processes: processes, Temperatures: now.Temperatures, MaxTemperature: maxTemperature, TemperatureAvailable: hasTemperature}
	analysis.Diagnoses = diagnoses(history, sample, cpuCount, processes, hasTemperature)
	active := make(map[string]bool, len(analysis.Diagnoses))
	for _, diagnosis := range analysis.Diagnoses {
		key := "diagnosis:" + diagnosis.Title
		active[key] = true
		if diagnosis.Severity != models.SeverityOK {
			history.SetCondition(key, true, models.Event{Timestamp: now.Timestamp, Severity: diagnosis.Severity, Message: diagnosis.Title})
		}
	}
	for _, title := range []string{"CPU pressure", "Elevated CPU usage", "Memory pressure", "Elevated memory usage", "Active swapping", "Load pressure", "High disk activity", "High temperature"} {
		history.SetCondition("diagnosis:"+title, active["diagnosis:"+title], models.Event{})
	}
	previous := history.Samples()
	if len(previous) >= 2 && sample.CPUUsage >= 75 && sample.CPUUsage-previous[len(previous)-2].CPUUsage >= 20 {
		history.SetCondition("cpu-rise", true, models.Event{Timestamp: now.Timestamp, Severity: models.SeverityWarning, Message: "CPU usage increased significantly"})
	} else {
		history.SetCondition("cpu-rise", false, models.Event{})
	}
	if len(previous) >= 2 && sample.TopProcess != "" && previous[len(previous)-2].TopProcess != "" && sample.TopProcess != previous[len(previous)-2].TopProcess && len(processes) > 0 && processes[0].CPUUsage >= 25 {
		history.SetCondition("top-process:"+sample.TopProcess, true, models.Event{Timestamp: now.Timestamp, Severity: models.SeverityInfo, Message: fmt.Sprintf("%s became the top CPU contributor", sample.TopProcess)})
	}
	analysis.Events = history.Events()
	return analysis
}

func diagnoses(history *History, sample models.Sample, cpuCount int, processes []models.ProcessStats, hasTemperature bool) []models.Diagnosis {
	const sustainedSamples = 3
	result := make([]models.Diagnosis, 0, 6)
	contributors := topContributors(processes)
	if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.CPUUsage >= 90 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityCritical, Title: "CPU pressure", Message: "CPU utilization has remained above 90% for several samples.", Contributors: contributors})
	} else if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.CPUUsage >= 75 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityWarning, Title: "Elevated CPU usage", Message: "CPU utilization has remained above 75% for several samples.", Contributors: contributors})
	}
	if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.MemoryUsage >= 90 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityCritical, Title: "Memory pressure", Message: "Memory use has remained above 90%; pressure is more likely if swapping is also active."})
	} else if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.MemoryUsage >= 80 && s.SwapPagesPerSec > 0 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityWarning, Title: "Elevated memory usage", Message: "Memory use is elevated and the system is actively swapping."})
	}
	if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.SwapPagesPerSec > 0 && s.MemoryUsage >= 80 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityWarning, Title: "Active swapping", Message: "Swap page activity has persisted alongside elevated memory use."})
	}
	if cpuCount > 0 && history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.Load1 >= float64(cpuCount) }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityWarning, Title: "Load pressure", Message: fmt.Sprintf("1-minute load has remained at or above %d logical CPUs; work may be queueing.", cpuCount), Contributors: contributors})
	}
	if history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.ReadBytesPerSec+s.WriteBytesPerSec >= 100*1024*1024 }) {
		result = append(result, models.Diagnosis{Severity: models.SeverityInfo, Title: "High disk activity", Message: "Disk throughput has remained high. This is activity, not proof that storage is the bottleneck."})
	}
	if hasTemperature && history.Sustained(sustainedSamples, func(s models.Sample) bool { return s.MaxTemperature >= 85 }) {
		severity := models.SeverityWarning
		if sample.MaxTemperature >= 95 {
			severity = models.SeverityCritical
		}
		result = append(result, models.Diagnosis{Severity: severity, Title: "High temperature", Message: fmt.Sprintf("A reported sensor has remained at %.0f°C or higher. No throttling is inferred without direct evidence.", sample.MaxTemperature)})
	}
	if len(result) == 0 {
		return []models.Diagnosis{{Severity: models.SeverityOK, Title: "No sustained pressure", Message: "No CPU, memory, load, disk, swap, or temperature pressure has persisted long enough to flag."}}
	}
	return result
}

func topContributors(processes []models.ProcessStats) []models.ProcessStats {
	limit := min(3, len(processes))
	result := make([]models.ProcessStats, 0, limit)
	for _, process := range processes[:limit] {
		if process.CPUUsage < 5 {
			break
		}
		result = append(result, process)
	}
	return result
}

func maximumTemperature(temperatures []models.Temperature) (float64, bool) {
	var maximum float64
	for _, temperature := range temperatures {
		if temperature.Celsius > maximum {
			maximum = temperature.Celsius
		}
	}
	return maximum, maximum > 0
}
