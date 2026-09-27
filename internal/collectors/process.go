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
	childs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var mem uint64

	processes := make([]models.ProcessStats, 0)

	for _, child := range childs {
		name := child.Name()

		if !child.IsDir() {
			continue
		}

		pid, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			continue
		}

		path := path.Join("/proc/", child.Name(), "status")
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				panic(err)
			}

			fields := strings.Fields(scanner.Text())
			switch fields[0] {
			case "VmRSS:":
				parsed, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					panic(err)
				}

				mem = parsed
			case "Name:":
				name = fields[1]
			}
		}

		process := models.ProcessStats{PID: pid, Name: name, Memory: mem}
		processes = append(processes, process)
	}

	return processes, nil
}
