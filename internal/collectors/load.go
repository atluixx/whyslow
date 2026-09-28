package collectors

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

func ReadLoadStats() (models.LoadStats, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return models.LoadStats{}, err
	}
	return parseLoadStats(data)
}

func parseLoadStats(data []byte) (models.LoadStats, error) {
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return models.LoadStats{}, fmt.Errorf("unexpected /proc/loadavg format")
	}
	parse := func(value string) (float64, error) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}

		return parsed, nil
	}

	values := make([]float64, 3)

	for i, v := range fields[:3] {
		parsed, err := parse(v)
		if err != nil {
			return models.LoadStats{}, err
		}
		values[i] = parsed
	}

	return models.LoadStats{
		Load1:  values[0],
		Load5:  values[1],
		Load15: values[2],
	}, nil
}
