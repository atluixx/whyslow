package models

type CPUStats struct {
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	Iowait    uint64
	IRQ       uint64
	SoftIRQ   uint64
	Steal     uint64
	Guest     uint64
	GuestNice uint64
}

// CPUActivity describes CPU use over one sampling interval.
type CPUActivity struct {
	Usage      float64
	Iowait     float64
	TotalTicks uint64
}
