package model

// Recommendation holds the computed recommendation for a single container.
type Recommendation struct {
	Cluster                     string   `json:"cluster"`
	Namespace                   string   `json:"namespace"`
	WorkloadKind                string   `json:"workload_kind"`
	WorkloadName                string   `json:"workload_name"`
	ContainerName               string   `json:"container_name"`
	Team                        string   `json:"team"`
	Service                     string   `json:"service"`
	Environment                 string   `json:"environment"`
	CurrentCPURequestMcores     float64  `json:"current_cpu_request_mcores"`
	CurrentMemoryRequestMiB     float64  `json:"current_memory_request_mib"`
	RecommendedCPURequestMcores float64  `json:"recommended_cpu_request_mcores"`
	RecommendedMemoryRequestMiB float64  `json:"recommended_memory_request_mib"`
	CPUStatus                   string   `json:"cpu_status"`
	MemoryStatus                string   `json:"memory_status"`
	Action                      string   `json:"action"`
	Risk                        string   `json:"risk"`
	Confidence                  string   `json:"confidence"`
	CPURequestDeltaMcores       float64  `json:"cpu_request_delta_mcores"`
	MemoryRequestDeltaMiB       float64  `json:"memory_request_delta_mib"`
	CPUReductionPercent         float64  `json:"cpu_reduction_percent"`
	MemoryReductionPercent      float64  `json:"memory_reduction_percent"`
	Reasons                     []string `json:"reasons"`
	Suggestions                 []string `json:"suggestions"`
}
