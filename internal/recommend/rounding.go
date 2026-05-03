package recommend

import "math"

const (
	MinCPUMcores  = 10.0
	MinMemoryMiB  = 32.0
	CPURoundTo    = 10.0
	MemoryRoundTo = 16.0
)

// RoundUpCPU rounds a CPU value (in mcores) up to the nearest 10m.
func RoundUpCPU(mcores float64) float64 {
	if mcores < MinCPUMcores {
		return MinCPUMcores
	}
	return math.Ceil(mcores/CPURoundTo) * CPURoundTo
}

// RoundUpMemory rounds a memory value (in MiB) up to the nearest 16Mi.
func RoundUpMemory(mib float64) float64 {
	if mib < MinMemoryMiB {
		return MinMemoryMiB
	}
	return math.Ceil(mib/MemoryRoundTo) * MemoryRoundTo
}
