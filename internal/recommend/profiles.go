package recommend

// Profile defines multipliers used to compute recommended requests.
type Profile struct {
	Name            string
	CPUP95Mult      float64
	CPUP99Mult      float64 // 0 means p99 is not used
	MemoryP95Mult   float64
	MemoryP99Mult   float64 // 0 means p99 is not used
	UseP99ForCPU    bool
	UseP99ForMemory bool
}

// Profiles maps profile names to their configuration.
var Profiles = map[string]Profile{
	"conservative": {
		Name:            "conservative",
		CPUP95Mult:      1.5,
		CPUP99Mult:      1.1,
		MemoryP95Mult:   1.3,
		MemoryP99Mult:   1.1,
		UseP99ForCPU:    true,
		UseP99ForMemory: true,
	},
	"balanced": {
		Name:            "balanced",
		CPUP95Mult:      1.25,
		CPUP99Mult:      1.0,
		MemoryP95Mult:   1.2,
		MemoryP99Mult:   1.0,
		UseP99ForCPU:    true,
		UseP99ForMemory: true,
	},
	"aggressive": {
		Name:            "aggressive",
		CPUP95Mult:      1.1,
		CPUP99Mult:      0,
		MemoryP95Mult:   1.1,
		MemoryP99Mult:   0,
		UseP99ForCPU:    false,
		UseP99ForMemory: false,
	},
}

// ValidProfiles lists all allowed profile names.
var ValidProfiles = []string{"conservative", "balanced", "aggressive"}

// IsValidProfile returns true if name is a known profile.
func IsValidProfile(name string) bool {
	_, ok := Profiles[name]
	return ok
}
