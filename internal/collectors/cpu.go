package collectors

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

func ReadCPUStats() (models.CPUStats, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return models.CPUStats{}, err
	}

	line, _, _ := strings.Cut(string(data), "\n")
	fields := strings.Fields(line)

	if len(fields) < 11 {
		return models.CPUStats{}, fmt.Errorf("unexpected /proc/stat format")
	}

	if fields[0] != "cpu" {
		return models.CPUStats{}, fmt.Errorf("unexpected /proc/stat CPU line")
	}

	values := make([]uint64, 10)

	for i, field := range fields[1:11] {
		values[i], err = strconv.ParseUint(field, 10, 64)
		if err != nil {
			return models.CPUStats{}, err
		}
	}

	return models.CPUStats{
		User:      values[0],
		Nice:      values[1],
		System:    values[2],
		Idle:      values[3],
		Iowait:    values[4],
		IRQ:       values[5],
		SoftIRQ:   values[6],
		Steal:     values[7],
		Guest:     values[8],
		GuestNice: values[9],
	}, nil
}
