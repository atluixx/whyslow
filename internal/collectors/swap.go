package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

func ReadSwapActivity() (models.SwapActivityStats, error) {
	file, err := os.Open("/proc/vmstat")
	if err != nil {
		return models.SwapActivityStats{}, err
	}
	defer file.Close()
	var stats models.SwapActivityStats
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || (fields[0] != "pswpin" && fields[0] != "pswpout") {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return models.SwapActivityStats{}, err
		}
		if fields[0] == "pswpin" {
			stats.PagesIn = value
		} else {
			stats.PagesOut = value
		}
	}
	if err := scanner.Err(); err != nil {
		return models.SwapActivityStats{}, err
	}
	return stats, nil
}
