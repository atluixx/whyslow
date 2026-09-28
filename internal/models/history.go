package models

import "time"

type Sample struct {
	Timestamp        time.Time
	CPUUsage         float64
	MemoryUsage      float64
	SwapUsage        float64
	Load1            float64
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	SwapPagesPerSec  float64
	MaxTemperature   float64
	TopProcess       string
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
	SeverityOK       Severity = "ok"
)

type Event struct {
	Timestamp time.Time
	Severity  Severity
	Message   string
}
