package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	cyan   = "\x1b[36m"
)

// Dashboard is the complete state rendered on each refresh.
type Dashboard struct {
	CPUUsage     float64
	Iowait       float64
	MemoryUsage  float64
	SwapUsage    float64
	Load         models.LoadStats
	Diagnoses    []models.Diagnosis
	Processes    []models.ProcessStats
	RefreshLabel string
}

// Render writes one complete terminal dashboard. It uses ANSI sequences only,
// avoiding a subprocess for every refresh.
func Render(out io.Writer, dashboard Dashboard, color bool) {
	paint := func(code, value string) string {
		if !color {
			return value
		}
		return code + value + reset
	}
	levelColor := func(level string) string {
		switch level {
		case "critical":
			return red
		case "warning":
			return yellow
		case "ok":
			return green
		default:
			return cyan
		}
	}
	fprint := func(format string, args ...any) { _, _ = fmt.Fprintf(out, format, args...) }
	fprint("\x1b[H\x1b[2J")
	fprint("%s\n", paint(bold+cyan, "whyslow — Linux performance dashboard"))
	fprint("Refresh: %s  •  Ctrl-C to quit\n\n", dashboard.RefreshLabel)
	fprint("CPU %6.2f%%   I/O wait %6.2f%%   Memory %6.2f%%   Swap %6.2f%%\n", dashboard.CPUUsage, dashboard.Iowait, dashboard.MemoryUsage, dashboard.SwapUsage)
	fprint("Load average  %.2f  %.2f  %.2f\n\n", dashboard.Load.Load1, dashboard.Load.Load5, dashboard.Load.Load15)
	fprint("%s\n", paint(bold, "Diagnosis"))
	for _, diagnosis := range dashboard.Diagnoses {
		fprint("%s %s — %s\n", paint(levelColor(diagnosis.Level)+bold, strings.ToUpper(diagnosis.Level)), diagnosis.Title, diagnosis.Details)
	}
	fprint("\n%s\n", paint(bold, "Top processes"))
	fprint("----------------------------------------------------------------------------\n")
	fprint("%-8s %-28s %8s %9s %12s\n", "PID", "COMMAND", "%CPU", "THREADS", "RSS")
	fprint("----------------------------------------------------------------------------\n")
	for _, process := range dashboard.Processes {
		fprint("%-8d %-28s %7.2f%% %9d %10.2f MB\n", process.PID, truncate(process.Name, 28), process.CPUUsage, process.Threads, float64(process.RSS)/(1024*1024))
	}
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
