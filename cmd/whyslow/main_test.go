package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	previous := Version
	Version = "0.1.0-test"
	t.Cleanup(func() { Version = previous })
	var output bytes.Buffer
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "whyslow version 0.1.0-test\n" {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestHelpCommand(t *testing.T) {
	var output bytes.Buffer
	if err := execute([]string{"help"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "install") || !strings.Contains(output.String(), "status") {
		t.Fatalf("unexpected help: %q", output.String())
	}
}

func TestParseMonitorConfig(t *testing.T) {
	var output bytes.Buffer
	config, err := parseMonitorConfig([]string{"--interval=2s", "--top=4", "--color=false"}, &output, false)
	if err != nil || config.Interval.String() != "2s" || config.Top != 4 || config.Color {
		t.Fatalf("unexpected config %#v, %v", config, err)
	}
	if _, err := parseMonitorConfig([]string{"--interval=0s"}, &output, true); err == nil {
		t.Fatal("expected invalid interval")
	}
}
