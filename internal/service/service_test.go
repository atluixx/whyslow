package service

import (
	"strings"
	"testing"
)

func TestUnitContents(t *testing.T) {
	unit := UnitContents()
	for _, expected := range []string{unitMarker, "ExecStart=/usr/local/bin/whyslow service --interval=5s", "DynamicUser=yes", "NoNewPrivileges=yes", "Restart=on-failure"} {
		if !strings.Contains(unit, expected) {
			t.Fatalf("unit is missing %q:\n%s", expected, unit)
		}
	}
}

func TestParsePropertiesAndDuration(t *testing.T) {
	properties := parseProperties("ActiveState=active\nMainPID=42\n")
	if properties["ActiveState"] != "active" || properties["MainPID"] != "42" {
		t.Fatalf("unexpected properties: %#v", properties)
	}
	if got := formatDuration(3723); got != "1h 2m 3s" {
		t.Fatalf("unexpected duration: %s", got)
	}
}
