package analyzers

import (
	"math"
	"testing"

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

func TestProcessUsageSkipsCounterReset(t *testing.T) {
	cpuOld := models.CPUStats{Idle: 100}
	cpuNow := models.CPUStats{Idle: 200}
	processes := ProcessUsageWithCPUCount(
		[]models.ProcessStats{{PID: 1, CPUTime: 50}},
		[]models.ProcessStats{{PID: 1, CPUTime: 10}, {PID: 2, CPUTime: 20}},
		cpuOld, cpuNow, 1,
	)
	if len(processes) != 0 {
		t.Fatalf("expected reset and new processes to be omitted, got %#v", processes)
	}
}

func TestDiagnose(t *testing.T) {
	diagnoses := Diagnose(
		models.CPUActivity{Usage: 90, Iowait: 20},
		models.MemoryStats{Total: 100, Available: 5, SwapTotal: 100, SwapFree: 80},
		models.LoadStats{Load1: 8}, 4,
		[]models.ProcessStats{{PID: 12, Name: "hog", CPUUsage: 50}},
	)
	if len(diagnoses) != 6 {
		t.Fatalf("expected six diagnoses, got %#v", diagnoses)
	}
}

func TestDiagnoseHealthy(t *testing.T) {
	diagnoses := Diagnose(models.CPUActivity{Usage: 10}, models.MemoryStats{Total: 100, Available: 80}, models.LoadStats{Load1: 1}, 4, nil)
	if len(diagnoses) != 1 || diagnoses[0].Level != "ok" {
		t.Fatalf("expected healthy diagnosis, got %#v", diagnoses)
	}
}
