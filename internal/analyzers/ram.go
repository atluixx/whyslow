package analyzers

import "github.com/atluixx/whyslow/internal/models"

func MemoryUsage(stats models.MemoryStats) float64 {
	used := stats.Total - stats.Available
	return float64(used) / float64(stats.Total) * 100
}
