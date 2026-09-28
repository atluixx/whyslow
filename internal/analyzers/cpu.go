package analyzers

import "github.com/atluixx/whyslow/internal/models"

func CPUUsage(old, now models.CPUStats) float64 {
	return CPUActivity(old, now).Usage
}

// CPUActivity calculates usage safely. A counter reset produces a zero value
// rather than an unsigned integer underflow.
func CPUActivity(old, now models.CPUStats) models.CPUActivity {
	oldValues := []uint64{old.User, old.Nice, old.System, old.Idle, old.Iowait, old.IRQ, old.SoftIRQ, old.Steal, old.Guest, old.GuestNice}
	nowValues := []uint64{now.User, now.Nice, now.System, now.Idle, now.Iowait, now.IRQ, now.SoftIRQ, now.Steal, now.Guest, now.GuestNice}
	deltas := make([]uint64, len(oldValues))
	for i := range oldValues {
		if nowValues[i] < oldValues[i] {
			return models.CPUActivity{}
		}
		deltas[i] = nowValues[i] - oldValues[i]
	}

	busy := deltas[0] + deltas[1] + deltas[2] + deltas[5] + deltas[6] + deltas[7]
	total := busy + deltas[3] + deltas[4]
	if total == 0 {
		return models.CPUActivity{}
	}
	return models.CPUActivity{Usage: float64(busy) / float64(total) * 100, Iowait: float64(deltas[4]) / float64(total) * 100, TotalTicks: total}
}
