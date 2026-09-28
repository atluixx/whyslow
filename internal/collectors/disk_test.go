package collectors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDiskStatsFiltersPartitions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diskstats")
	data := "8 0 sda 10 0 20 0 30 0 40 0 0 50 0\n8 1 sda1 99 0 99 0 99 0 99 0 0 99 0\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	stats, err := parseDiskStats(file, map[string]bool{"sda": true})
	if err != nil || stats.ReadsCompleted != 10 || stats.SectorsRead != 20 || stats.WritesCompleted != 30 || stats.SectorsWritten != 40 || stats.IOTimeMillis != 50 {
		t.Fatalf("unexpected stats: %#v, %v", stats, err)
	}
}
