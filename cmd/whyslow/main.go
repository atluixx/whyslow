package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/atluixx/whyslow/internal/analyzers"
	"github.com/atluixx/whyslow/internal/collectors"
	"github.com/atluixx/whyslow/internal/ui"
)

type config struct {
	interval time.Duration
	top      int
	color    bool
}

func main() {
	configuration, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "whyslow:", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer ui.Restore(os.Stdout)
	if err := run(ctx, configuration); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "whyslow:", err)
		os.Exit(1)
	}
}

func parseConfig() (config, error) {
	configuration := config{}
	flag.DurationVar(&configuration.interval, "interval", time.Second, "refresh interval (for example: 500ms or 2s)")
	flag.IntVar(&configuration.top, "top", 10, "number of processes to display")
	flag.BoolVar(&configuration.color, "color", true, "use ANSI color")
	flag.Parse()
	if configuration.interval <= 0 {
		return config{}, errors.New("interval must be greater than zero")
	}
	if configuration.top <= 0 {
		return config{}, errors.New("top must be greater than zero")
	}
	return configuration, nil
}

func run(ctx context.Context, configuration config) error {
	oldSnapshot, err := collectors.CollectSnapshot()
	if err != nil {
		return fmt.Errorf("collect initial snapshot: %w", err)
	}
	history := analyzers.NewHistory(120)
	ticker := time.NewTicker(configuration.interval)
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
		analysis := analyzers.Analyze(oldSnapshot, nowSnapshot, history, runtime.NumCPU())
		processes := analysis.Processes
		if len(processes) > configuration.top {
			processes = processes[:configuration.top]
		}
		ui.Render(os.Stdout, ui.Dashboard{
			CPUUsage: analysis.Activity.Usage, Iowait: analysis.Activity.Iowait, MemoryUsage: analysis.MemoryUsage, SwapUsed: analysis.SwapUsed, SwapUsage: analysis.SwapUsage,
			Load: nowSnapshot.Load, Disk: analysis.Disk, MaxTemperature: analysis.MaxTemperature, TemperatureAvailable: analysis.TemperatureAvailable,
			Processes: processes, RefreshLabel: configuration.interval.String(), UpdatedAt: nowSnapshot.Timestamp, Diagnoses: analysis.Diagnoses, Events: analysis.Events,
		}, configuration.color)
		oldSnapshot = nowSnapshot
	}
}
