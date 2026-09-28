package collectors

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

// ReadTemperatures returns any readable millidegree-Celsius sensors. Missing
// or inaccessible sensors are normal and are silently omitted.
func ReadTemperatures() []models.Temperature {
	paths := make([]string, 0)
	thermal, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	hwmon, _ := filepath.Glob("/sys/class/hwmon/hwmon*/temp*_input")
	paths = append(paths, thermal...)
	paths = append(paths, hwmon...)
	result := make([]models.Temperature, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil || value <= 0 {
			continue
		}
		name := filepath.Base(filepath.Dir(path))
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_input") {
			name = strings.TrimSuffix(base, "_input")
		}
		result = append(result, models.Temperature{Name: name, Celsius: value / 1000})
	}
	return result
}
