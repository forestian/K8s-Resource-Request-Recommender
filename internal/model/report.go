package model

// Report is the top-level output of a recommend run.
type Report struct {
	Profile         string           `json:"profile"`
	InputFile       string           `json:"input_file"`
	TotalItems      int              `json:"total_items"`
	Summary         Summary          `json:"summary"`
	Recommendations []Recommendation `json:"recommendations"`
	Warnings        []string         `json:"warnings"`
}

// Summary aggregates counts and totals across all recommendations.
type Summary struct {
	SetRequests                 int     `json:"set_requests"`
	ReduceRequests              int     `json:"reduce_requests"`
	IncreaseRequests            int     `json:"increase_requests"`
	Review                      int     `json:"review"`
	NoChange                    int     `json:"no_change"`
	HighRisk                    int     `json:"high_risk"`
	MediumRisk                  int     `json:"medium_risk"`
	LowRisk                     int     `json:"low_risk"`
	PotentialCPUReductionMcores float64 `json:"potential_cpu_reduction_mcores"`
	PotentialMemoryReductionMiB float64 `json:"potential_memory_reduction_mib"`
}
