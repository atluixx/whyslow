package analyzers

import (
	"time"

	"github.com/atluixx/whyslow/internal/models"
)

const diskSectorBytes = 512

type DiskActivity struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	IOBusyPercent    float64
}

func DiskUsage(old, now models.DiskStats, interval time.Duration) DiskActivity {
	if interval <= 0 || now.SectorsRead < old.SectorsRead || now.SectorsWritten < old.SectorsWritten || now.IOTimeMillis < old.IOTimeMillis {
		return DiskActivity{}
	}
	seconds := interval.Seconds()
	return DiskActivity{
		ReadBytesPerSec:  float64((now.SectorsRead-old.SectorsRead)*diskSectorBytes) / seconds,
		WriteBytesPerSec: float64((now.SectorsWritten-old.SectorsWritten)*diskSectorBytes) / seconds,
		IOBusyPercent:    min(float64(now.IOTimeMillis-old.IOTimeMillis)/seconds/10, 100),
	}
}

func swapPagesPerSecond(old, now models.SwapActivityStats, interval time.Duration) float64 {
	if interval <= 0 || now.PagesIn < old.PagesIn || now.PagesOut < old.PagesOut {
		return 0
	}
	return float64(now.PagesIn-old.PagesIn+now.PagesOut-old.PagesOut) / interval.Seconds()
}
