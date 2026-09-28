package analyzers

import (
	"math"
	"testing"
	"time"

	"github.com/atluixx/whyslow/internal/models"
)

func TestCPUActivity(t *testing.T) {
	old := models.CPUStats{User: 10, System: 10, Idle: 80, Iowait: 5}
	now := models.CPUStats{User: 30, System: 20, Idle: 130, Iowait: 10}
	activity := CPUActivity(old, now)
	if activity.TotalTicks != 85 || math.Abs(activity.Usage-35.294) > .01 || math.Abs(activity.Iowait-5.882) > .01 {
		t.Fatalf("unexpected activity: %#v", activity)
	}
}

func TestCPUActivityCounterReset(t *testing.T) {
	activity := CPUActivity(models.CPUStats{User: 20, Iowait: 10}, models.CPUStats{User: 21, Iowait: 1})
	if activity != (models.CPUActivity{}) {
		t.Fatalf("counter reset must produce zero activity: %#v", activity)
	}
}

func TestMemoryAndSwapUsageInvalidValues(t *testing.T) {
	if MemoryUsage(models.MemoryStats{}) != 0 || SwapUsage(models.MemoryStats{}) != 0 {
		t.Fatal("zero totals must not produce NaN")
	}
	if MemoryUsage(models.MemoryStats{Total: 10, Available: 11}) != 0 || SwapUsage(models.MemoryStats{SwapTotal: 10, SwapFree: 11}) != 0 {
		t.Fatal("invalid available values must not underflow")
	}
}

func TestSwapUsage(t *testing.T) {
	if got := SwapUsage(models.MemoryStats{SwapTotal: 100, SwapFree: 25}); got != 75 {
		t.Fatalf("got %.2f, want 75", got)
	}
	if got := SwapUsed(models.MemoryStats{SwapTotal: 100, SwapFree: 25}); got != 75 {
		t.Fatalf("got %d, want 75", got)
	}
}

func TestProcessUsageSkipsCounterReset(t *testing.T) {
	processes := ProcessUsageWithCPUCount([]models.ProcessStats{{PID: 1, CPUTime: 50}}, []models.ProcessStats{{PID: 1, CPUTime: 10}, {PID: 2, CPUTime: 20}}, models.CPUStats{Idle: 100}, models.CPUStats{Idle: 200}, 1)
	if len(processes) != 0 {
		t.Fatalf("expected reset and new processes to be omitted, got %#v", processes)
	}
}

func TestDiskUsage(t *testing.T) {
	activity := DiskUsage(models.DiskStats{SectorsRead: 10, SectorsWritten: 20, IOTimeMillis: 100}, models.DiskStats{SectorsRead: 210, SectorsWritten: 120, IOTimeMillis: 600}, time.Second)
	if activity.ReadBytesPerSec != 200*diskSectorBytes || activity.WriteBytesPerSec != 100*diskSectorBytes || activity.IOBusyPercent != 50 {
		t.Fatalf("unexpected disk activity: %#v", activity)
	}
	if got := DiskUsage(models.DiskStats{SectorsRead: 100}, models.DiskStats{SectorsRead: 1}, time.Second); got != (DiskActivity{}) {
		t.Fatalf("counter reset must be ignored: %#v", got)
	}
}

func TestDiagnosisRequiresSustainedPressure(t *testing.T) {
	history := NewHistory(10)
	old := models.Snapshot{Timestamp: time.Unix(0, 0), CPU: models.CPUStats{Idle: 100}}
	for i := 1; i <= 2; i++ {
		now := models.Snapshot{Timestamp: time.Unix(int64(i), 0), CPU: models.CPUStats{User: uint64(i * 100), Idle: 100}}
		analysis := Analyze(old, now, history, 1)
		if analysis.Diagnoses[0].Severity != models.SeverityOK {
			t.Fatalf("sample %d should not yet diagnose: %#v", i, analysis.Diagnoses)
		}
		old = now
	}
	old.Processes = []models.ProcessStats{{PID: 1, Name: "hog", CPUTime: 200}}
	now := models.Snapshot{Timestamp: time.Unix(3, 0), CPU: models.CPUStats{User: 300, Idle: 100}, Processes: []models.ProcessStats{{PID: 1, Name: "hog", CPUTime: 300}}}
	analysis := Analyze(old, now, history, 1)
	if analysis.Diagnoses[0].Title != "CPU pressure" || len(analysis.Diagnoses[0].Contributors) != 1 {
		t.Fatalf("expected sustained CPU diagnosis with contributor: %#v", analysis.Diagnoses)
	}
}

func TestEventDeduplication(t *testing.T) {
	history := NewHistory(4)
	event := models.Event{Timestamp: time.Unix(1, 0), Severity: models.SeverityWarning, Message: "CPU pressure"}
	if !history.SetCondition("cpu", true, event) || history.SetCondition("cpu", true, event) || len(history.Events()) != 1 {
		t.Fatalf("expected one event, got %#v", history.Events())
	}
	history.SetCondition("cpu", false, models.Event{})
	if !history.SetCondition("cpu", true, event) || len(history.Events()) != 2 {
		t.Fatalf("expected event after condition recovery, got %#v", history.Events())
	}
}

func TestMemoryDiagnosisThreshold(t *testing.T) {
	history := NewHistory(4)
	for i := 0; i < 3; i++ {
		history.AddSample(models.Sample{MemoryUsage: 91})
	}
	diagnoses := diagnoses(history, models.Sample{MemoryUsage: 91}, 4, nil, false)
	if len(diagnoses) != 1 || diagnoses[0].Title != "Memory pressure" || diagnoses[0].Severity != models.SeverityCritical {
		t.Fatalf("unexpected memory diagnosis: %#v", diagnoses)
	}
}
