package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

func ReadMemoryStats() (models.MemoryStats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return models.MemoryStats{}, err
	}
	defer file.Close()

	var stats models.MemoryStats
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return models.MemoryStats{}, err
		}

		value *= 1024

		switch fields[0] {
		case "MemTotal:":
			stats.Total = value
		case "MemAvailable:":
			stats.Available = value
		case "SwapTotal:":
			stats.SwapTotal = value
		case "SwapFree:":
			stats.SwapFree = value
		}
	}

	if err := scanner.Err(); err != nil {
		return models.MemoryStats{}, err
	}

	return stats, nil
}
