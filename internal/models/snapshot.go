package models

import "time"

// Snapshot is one point-in-time view of the system collected from Linux procfs
// and sysfs. Counters remain cumulative; analyzers calculate deltas.
type Snapshot struct {
	Timestamp    time.Time
	CPU          CPUStats
	Memory       MemoryStats
	Load         LoadStats
	Processes    []ProcessStats
	Disk         DiskStats
	SwapActivity SwapActivityStats
	Temperatures []Temperature
}

type DiskStats struct {
	ReadsCompleted  uint64
	SectorsRead     uint64
	WritesCompleted uint64
	SectorsWritten  uint64
	IOTimeMillis    uint64
}

// SwapActivityStats contains cumulative page-ins and page-outs from vmstat.
type SwapActivityStats struct {
	PagesIn  uint64
	PagesOut uint64
}

type Temperature struct {
	Name    string
	Celsius float64
}
