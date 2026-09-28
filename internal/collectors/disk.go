package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

// ReadDiskStats aggregates whole-device counters from /proc/diskstats.
// Partitions are excluded by accepting only devices represented in /sys/block.
func ReadDiskStats() (models.DiskStats, error) {
	devices, err := os.ReadDir("/sys/block")
	if err != nil {
		return models.DiskStats{}, err
	}
	whole := make(map[string]bool, len(devices))
	for _, device := range devices {
		name := device.Name()
		if !strings.HasPrefix(name, "loop") && !strings.HasPrefix(name, "ram") {
			whole[name] = true
		}
	}
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return models.DiskStats{}, err
	}
	defer file.Close()
	return parseDiskStats(file, whole)
}

func parseDiskStats(file *os.File, whole map[string]bool) (models.DiskStats, error) {
	var result models.DiskStats
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 || !whole[fields[2]] {
			continue
		}
		values := []struct {
			index  int
			target *uint64
		}{{3, &result.ReadsCompleted}, {5, &result.SectorsRead}, {7, &result.WritesCompleted}, {9, &result.SectorsWritten}, {12, &result.IOTimeMillis}}
		for _, value := range values {
			parsed, err := strconv.ParseUint(fields[value.index], 10, 64)
			if err != nil {
				return models.DiskStats{}, err
			}
			*value.target += parsed
		}
	}
	return result, scanner.Err()
}
