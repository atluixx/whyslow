package analyzers

import "github.com/atluixx/whyslow/internal/models"

func MemoryUsage(stats models.MemoryStats) float64 {
	if stats.Total == 0 || stats.Available > stats.Total {
		return 0
	}
	used := stats.Total - stats.Available
	return float64(used) / float64(stats.Total) * 100
}

func SwapUsage(stats models.MemoryStats) float64 {
	if stats.SwapTotal == 0 {
		return 0
	}
	return float64(SwapUsed(stats)) / float64(stats.SwapTotal) * 100
}

func SwapUsed(stats models.MemoryStats) uint64 {
	if stats.SwapFree > stats.SwapTotal {
		return 0
	}
	return stats.SwapTotal - stats.SwapFree
}
