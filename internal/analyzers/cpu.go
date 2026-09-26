package analyzers

import "github.com/atluixx/whyslow/internal/models"

func CPUUsage(old, now models.CPUStats) float64 {
	user := now.User - old.User
	nice := now.Nice - old.Nice
	system := now.System - old.System
	idle := now.Idle - old.Idle
	iowait := now.Iowait - old.Iowait
	irq := now.IRQ - old.IRQ
	softirq := now.SoftIRQ - old.SoftIRQ
	steal := now.Steal - old.Steal

	busy := user + nice + system + irq + softirq + steal
	total := busy + idle + iowait

	if total == 0 {
		return 0
	}

	if now.User < old.User ||
		now.Nice < old.Nice ||
		now.System < old.System ||
		now.Idle < old.Idle {
		return 0
	}

	return float64(busy) / float64(total) * 100
}
