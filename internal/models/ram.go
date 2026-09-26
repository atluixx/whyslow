package models

type MemoryStats struct {
	Total     uint64
	Available uint64
	SwapTotal uint64
	SwapFree  uint64
}
