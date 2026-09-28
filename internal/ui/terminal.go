package ui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/atluixx/whyslow/internal/models"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	cyan   = "\x1b[36m"
	dim    = "\x1b[2m"
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
	UpdatedAt    time.Time
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
	metricColor := func(value, warning, critical float64) string {
		if value >= critical {
			return red
		}
		if value >= warning {
			return yellow
		}
		return green
	}
	metric := func(label string, value, warning, critical float64) string {
		text := fmt.Sprintf("%-8s %6.1f%% %s", label, value, bar(value, 12))
		return paint(metricColor(value, warning, critical)+bold, text)
	}
	updatedAt := dashboard.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	fprint := func(format string, args ...any) { _, _ = fmt.Fprintf(out, format, args...) }
	fprint("\x1b[?25l\x1b[H\x1b[2J")
	fprint("%s\n", paint(bold+cyan, "╭─ whyslow")+paint(dim, "  Linux performance monitor")+paint(green+bold, "  ● LIVE")+paint(dim, "  "+updatedAt.Format("15:04:05")+"  refresh "+dashboard.RefreshLabel))
	fprint("%s\n", paint(dim, "╰────────────────────────────────────────────────────────────────────────────"))
	fprint("%s\n", paint(bold, "SYSTEM HEALTH"))
	fprint("  %s   %s\n", metric("CPU", dashboard.CPUUsage, 70, 85), metric("MEMORY", dashboard.MemoryUsage, 75, 90))
	fprint("  %s   %s\n", metric("I/O WAIT", dashboard.Iowait, 8, 15), metric("SWAP", dashboard.SwapUsage, 5, 25))
	fprint("  %s\n\n", paint(dim, fmt.Sprintf("Load average    1m %-6.2f  5m %-6.2f  15m %-6.2f", dashboard.Load.Load1, dashboard.Load.Load5, dashboard.Load.Load15)))
	fprint("%s\n", paint(bold, "DIAGNOSIS"))
	for _, diagnosis := range dashboard.Diagnoses {
		badge := " " + strings.ToUpper(diagnosis.Level) + " "
		fprint("  %s %s\n    %s\n", paint(levelColor(diagnosis.Level)+bold, badge), paint(bold, diagnosis.Title), paint(dim, diagnosis.Details))
	}
	fprint("\n%s\n", paint(bold, "TOP PROCESSES")+paint(dim, "  ranked by CPU usage"))
	fprint("  %-7s %-26s %8s %-12s %8s %11s\n", "PID", "COMMAND", "CPU", "ACTIVITY", "THREADS", "RSS")
	fprint("  %s\n", paint(dim, "────────────────────────────────────────────────────────────────────────────"))
	for _, process := range dashboard.Processes {
		fprint("  %-7d %-26s %7.2f%% %-12s %8d %9.2f MB\n", process.PID, truncate(process.Name, 26), process.CPUUsage, paint(metricColor(process.CPUUsage, 25, 75), bar(process.CPUUsage, 10)), process.Threads, float64(process.RSS)/(1024*1024))
	}
	fprint("\n%s\n", paint(dim, "Ctrl-C to exit  •  "+strconv.Itoa(len(dashboard.Processes))+" processes shown"))
}

// Restore makes the cursor visible after the application exits.
func Restore(out io.Writer) {
	_, _ = fmt.Fprint(out, "\x1b[0m\x1b[?25h\n")
}

func bar(value float64, width int) string {
	filled := int(value / 100 * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("·", width-filled) + "]"
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
