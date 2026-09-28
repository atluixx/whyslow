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
	oldCPU, err := collectors.ReadCPUStats()
	if err != nil {
		return fmt.Errorf("read CPU statistics: %w", err)
	}
	oldProcesses, err := collectors.ReadProcesses()
	if err != nil {
		return fmt.Errorf("read process statistics: %w", err)
	}
	ticker := time.NewTicker(configuration.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		memory, err := collectors.ReadMemoryStats()
		if err != nil {
			return fmt.Errorf("read memory statistics: %w", err)
		}
		load, err := collectors.ReadLoadStats()
		if err != nil {
			return fmt.Errorf("read load statistics: %w", err)
		}
		nowCPU, err := collectors.ReadCPUStats()
		if err != nil {
			return fmt.Errorf("read CPU statistics: %w", err)
		}
		nowProcesses, err := collectors.ReadProcesses()
		if err != nil {
			return fmt.Errorf("read process statistics: %w", err)
		}
		activity := analyzers.CPUActivity(oldCPU, nowCPU)
		processes := analyzers.ProcessUsage(oldProcesses, nowProcesses, oldCPU, nowCPU)
		if len(processes) > configuration.top {
			processes = processes[:configuration.top]
		}
		ui.Render(os.Stdout, ui.Dashboard{
			CPUUsage: activity.Usage, Iowait: activity.Iowait, MemoryUsage: analyzers.MemoryUsage(memory), SwapUsage: analyzers.SwapUsage(memory),
			Load: load, Processes: processes, RefreshLabel: configuration.interval.String(),
			UpdatedAt: time.Now(),
			Diagnoses: analyzers.Diagnose(activity, memory, load, runtime.NumCPU(), processes),
		}, configuration.color)
		oldCPU, oldProcesses = nowCPU, nowProcesses
	}
}
