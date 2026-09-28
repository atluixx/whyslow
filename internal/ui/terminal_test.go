package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/atluixx/whyslow/internal/models"
)

func TestRenderWithoutColor(t *testing.T) {
	var output bytes.Buffer
	Render(&output, Dashboard{RefreshLabel: "1s", UpdatedAt: time.Date(2026, 9, 28, 10, 30, 0, 0, time.UTC), Diagnoses: []models.Diagnosis{{Severity: models.SeverityOK, Title: "Healthy", Message: "All clear"}}}, false)
	if strings.Contains(output.String(), "\x1b[32m") || !strings.Contains(output.String(), "SYSTEM HEALTH") || !strings.Contains(output.String(), "DIAGNOSIS") || !strings.Contains(output.String(), "Healthy") || !strings.Contains(output.String(), "10:30:00") {
		t.Fatalf("unexpected rendered dashboard: %q", output.String())
	}
}

func TestBar(t *testing.T) {
	if got := bar(50, 4); got != "[██··]" {
		t.Fatalf("unexpected bar: %q", got)
	}
	if got := bar(250, 4); got != "[████]" {
		t.Fatalf("bar should clamp high values: %q", got)
	}
}
