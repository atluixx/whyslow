package collectors

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/atluixx/whyslow/internal/models"
)

func ReadProcesses() ([]models.ProcessStats, error) {
	children, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	processes := make([]models.ProcessStats, 0)

	for _, child := range children {
		name := child.Name()

		if !child.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(name)
		if err != nil {
			continue
		}

		statusPath := path.Join("/proc", name, "status")
		statusFile, err := os.Open(statusPath)
		if err != nil {
			continue
		}

		var processName string
		scanner := bufio.NewScanner(statusFile)

		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())

			if len(fields) < 2 {
				continue
			}

			if fields[0] == "Name:" {
				processName = fields[1]
				break
			}
		}

		scannerErr := scanner.Err()
		statusFile.Close()

		if scannerErr != nil {
			continue
		}

		if processName == "" {
			continue
		}

		statPath := path.Join("/proc", name, "stat")
		statBuffer, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		statString := string(statBuffer)

		endComm := strings.LastIndex(statString, ")")
		if endComm == -1 {
			continue
		}

		fields := strings.Fields(statString[endComm+1:])

		if len(fields) <= 21 {
			continue
		}

		utime, err := strconv.ParseUint(fields[11], 10, 64)
		if err != nil {
			continue
		}

		stime, err := strconv.ParseUint(fields[12], 10, 64)
		if err != nil {
			continue
		}

		threads, err := strconv.ParseUint(fields[17], 10, 64)
		if err != nil {
			continue
		}

		rss, err := strconv.ParseUint(fields[21], 10, 64)
		if err != nil {
			continue
		}

		cpuTime := utime + stime
		rssBytes := rss * uint64(os.Getpagesize())

		processes = append(processes, models.ProcessStats{
			PID:     pid,
			Name:    processName,
			CPUTime: cpuTime,
			Threads: threads,
			RSS:     rssBytes,
		})
	}

	return processes, nil
}
