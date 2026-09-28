package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/atluixx/whyslow/internal/models"
)

func TestRenderWithoutColor(t *testing.T) {
	var output bytes.Buffer
	Render(&output, Dashboard{RefreshLabel: "1s", Diagnoses: []models.Diagnosis{{Level: "ok", Title: "Healthy", Details: "All clear"}}}, false)
	if strings.Contains(output.String(), "\x1b[32m") || !strings.Contains(output.String(), "Diagnosis") || !strings.Contains(output.String(), "Healthy") {
		t.Fatalf("unexpected rendered dashboard: %q", output.String())
	}
}
