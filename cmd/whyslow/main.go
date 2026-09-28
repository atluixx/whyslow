package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/atluixx/whyslow/internal/monitor"
	"github.com/atluixx/whyslow/internal/service"
	"github.com/atluixx/whyslow/internal/ui"
)

// Version is replaced at build time with: -ldflags "-X main.Version=0.1.0".
var Version = "dev"

func main() {
	if err := execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "whyslow:", err)
		os.Exit(1)
	}
}

func execute(args []string, output io.Writer) error {
	command, arguments := "run", args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, arguments = args[0], args[1:]
	}
	switch command {
	case "run":
		config, err := parseMonitorConfig(arguments, output, false)
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		if err != nil {
			return err
		}
		return runInteractive(config)
	case "service":
		config, err := parseMonitorConfig(arguments, output, true)
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		if err != nil {
			return err
		}
		return runService(config)
	case "version":
		if len(arguments) != 0 {
			return fmt.Errorf("version accepts no arguments")
		}
		_, _ = fmt.Fprintf(output, "whyslow version %s\n", Version)
		return nil
	case "help", "--help", "-h":
		printHelp(output)
		return nil
	case "install":
		if err := noArguments(command, arguments); err != nil {
			return err
		}
		if err := service.Install(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(output, "whyslow service installed and enabled; use `whyslow start` to start it.")
		return nil
	case "uninstall":
		if err := noArguments(command, arguments); err != nil {
			return err
		}
		if err := service.Uninstall(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(output, "whyslow service uninstalled.")
		return nil
	case "start", "stop", "restart":
		if err := noArguments(command, arguments); err != nil {
			return err
		}
		var err error
		switch command {
		case "start":
			err = service.Start()
		case "stop":
			err = service.Stop()
		default:
			err = service.Restart()
		}
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(output, "whyslow service %sed.\n", command)
		return nil
	case "status":
		if err := noArguments(command, arguments); err != nil {
			return err
		}
		status, err := service.GetStatus()
		if err != nil {
			return err
		}
		if !status.Installed {
			_, _ = fmt.Fprintln(output, "whyslow service: not installed")
			return nil
		}
		_, _ = fmt.Fprintf(output, "whyslow service: %s\n", status.State)
		if status.PID > 0 {
			_, _ = fmt.Fprintf(output, "PID:             %d\n", status.PID)
		}
		if status.Uptime != "" {
			_, _ = fmt.Fprintf(output, "Uptime:          %s\n", status.Uptime)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q; run `whyslow help`", command)
	}
}

func parseMonitorConfig(args []string, output io.Writer, serviceMode bool) (monitor.Config, error) {
	defaults := monitor.Config{Interval: time.Second, Top: 10, Color: true}
	name := "run"
	if serviceMode {
		defaults.Interval, defaults.Color, name = 5*time.Second, false, "service"
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(output)
	flags.DurationVar(&defaults.Interval, "interval", defaults.Interval, "sampling interval (for example: 500ms or 5s)")
	if !serviceMode {
		flags.IntVar(&defaults.Top, "top", defaults.Top, "number of processes to display")
		flags.BoolVar(&defaults.Color, "color", defaults.Color, "use ANSI color")
	}
	if err := flags.Parse(args); err != nil {
		return monitor.Config{}, err
	}
	if flags.NArg() != 0 {
		return monitor.Config{}, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	return defaults, defaults.Validate()
}

func runInteractive(config monitor.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer ui.Restore(os.Stdout)
	err := monitor.RunInteractive(ctx, os.Stdout, config)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func runService(config monitor.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := monitor.RunService(ctx, config)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func noArguments(command string, arguments []string) error {
	if len(arguments) != 0 {
		return fmt.Errorf("%s accepts no arguments", command)
	}
	return nil
}

func printHelp(output io.Writer) {
	_, _ = fmt.Fprint(output, `whyslow — Linux performance diagnostic tool

Usage:
  whyslow [run] [--interval DURATION] [--top N] [--color=false]
  whyslow <command>

Commands:
  run        Run the interactive diagnostic dashboard (default)
  status     Show the installed systemd service status
  start      Start the installed systemd service
  stop       Stop the installed systemd service
  restart    Restart the installed systemd service
  install    Install and enable the systemd service (requires root)
  uninstall  Stop and remove files created by install (requires root)
  version    Print the build version
  help       Show this help

`)
}
