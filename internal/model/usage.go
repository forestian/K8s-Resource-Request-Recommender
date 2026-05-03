package model

// UsageFile wraps a list of usage items from JSON input.
type UsageFile struct {
	Items []UsageItem `json:"items"`
}

// UsageItem holds pre-collected resource usage statistics for one container.
type UsageItem struct {
	Cluster                 string  `json:"cluster"`
	Namespace               string  `json:"namespace"`
	WorkloadKind            string  `json:"workload_kind"`
	WorkloadName            string  `json:"workload_name"`
	ContainerName           string  `json:"container_name"`
	Team                    string  `json:"team"`
	Service                 string  `json:"service"`
	Environment             string  `json:"environment"`
	Window                  string  `json:"window"`
	CurrentCPURequestMcores float64 `json:"current_cpu_request_mcores"`
	CurrentMemoryRequestMiB float64 `json:"current_memory_request_mib"`
	CurrentCPULimitMcores   float64 `json:"current_cpu_limit_mcores"`
	CurrentMemoryLimitMiB   float64 `json:"current_memory_limit_mib"`
	CPUP50Mcores            float64 `json:"cpu_p50_mcores"`
	CPUP95Mcores            float64 `json:"cpu_p95_mcores"`
	CPUP99Mcores            float64 `json:"cpu_p99_mcores"`
	CPUMaxMcores            float64 `json:"cpu_max_mcores"`
	MemoryP50MiB            float64 `json:"memory_p50_mib"`
	MemoryP95MiB            float64 `json:"memory_p95_mib"`
	MemoryP99MiB            float64 `json:"memory_p99_mib"`
	MemoryMaxMiB            float64 `json:"memory_max_mib"`
	Samples                 int     `json:"samples"`
}

// Key returns a unique string identifying this item for duplicate detection.
func (u *UsageItem) Key() string {
	return u.Cluster + "/" + u.Namespace + "/" + u.WorkloadKind + "/" + u.WorkloadName + "/" + u.ContainerName
}

// ApplyDefaults fills in optional fields with their default values.
func (u *UsageItem) ApplyDefaults() {
	if u.Cluster == "" {
		u.Cluster = "unknown"
	}
	if u.Team == "" {
		u.Team = "unknown"
	}
	if u.Service == "" {
		u.Service = u.WorkloadName
	}
	if u.Environment == "" {
		u.Environment = "unknown"
	}
	if u.Window == "" {
		u.Window = "unknown"
	}
}
