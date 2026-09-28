package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/atluixx/whyslow/internal/analyzers"
	"github.com/atluixx/whyslow/internal/collectors"
	"github.com/atluixx/whyslow/internal/models"
	"github.com/atluixx/whyslow/internal/ui"
)

type Config struct {
	Interval time.Duration
	Top      int
	Color    bool
}

func (c Config) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}
	if c.Top <= 0 {
		return fmt.Errorf("top must be greater than zero")
	}
	return nil
}

func RunInteractive(ctx context.Context, output io.Writer, config Config) error {
	return run(ctx, config, func(snapshot models.Snapshot, analysis analyzers.Analysis) {
		processes := analysis.Processes
		if len(processes) > config.Top {
			processes = processes[:config.Top]
		}
		ui.Render(output, ui.Dashboard{
			CPUUsage: analysis.Activity.Usage, Iowait: analysis.Activity.Iowait, MemoryUsage: analysis.MemoryUsage, SwapUsed: analysis.SwapUsed, SwapUsage: analysis.SwapUsage,
			Load: snapshot.Load, Disk: analysis.Disk, MaxTemperature: analysis.MaxTemperature, TemperatureAvailable: analysis.TemperatureAvailable,
			Processes: processes, RefreshLabel: config.Interval.String(), UpdatedAt: snapshot.Timestamp, Diagnoses: analysis.Diagnoses, Events: analysis.Events,
		}, config.Color)
	})
}

// RunService is deliberately a normal foreground process. systemd supplies
// supervision, background execution, restarts, and journald capture.
func RunService(ctx context.Context, config Config) error {
	seenEvents := 0
	return run(ctx, config, func(_ models.Snapshot, analysis analyzers.Analysis) {
		events := analysis.Events
		if seenEvents > len(events) {
			seenEvents = 0
		}
		for _, event := range events[seenEvents:] {
			log.Printf("[%s] %s", event.Severity, event.Message)
		}
		seenEvents = len(events)
	})
}

func run(ctx context.Context, config Config, consume func(models.Snapshot, analyzers.Analysis)) error {
	if err := config.Validate(); err != nil {
		return err
	}
	oldSnapshot, err := collectors.CollectSnapshot()
	if err != nil {
		return fmt.Errorf("collect initial snapshot: %w", err)
	}
	history := analyzers.NewHistory(120)
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		nowSnapshot, err := collectors.CollectSnapshot()
		if err != nil {
			return fmt.Errorf("collect snapshot: %w", err)
		}
		consume(nowSnapshot, analyzers.Analyze(oldSnapshot, nowSnapshot, history, runtime.NumCPU()))
		oldSnapshot = nowSnapshot
	}
}
