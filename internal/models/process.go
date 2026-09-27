package models

type ProcessStats struct {
	PID      int
	Name     string
	CPUTime  uint64
	Threads  uint64
	RSS      uint64
	CPUUsage float64
}
