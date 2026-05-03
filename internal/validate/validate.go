package validate

import (
	"fmt"

	"github.com/krr-lite/krr-lite/internal/model"
)

// Result holds the outcome of validating a set of usage items.
type Result struct {
	Errors   []string
	Warnings []string
}

// OK returns true when there are no validation errors.
func (r *Result) OK() bool {
	return len(r.Errors) == 0
}

// Validate checks all items for correctness and returns errors and warnings.
func Validate(items []model.UsageItem, minSamples int) Result {
	var res Result
	seen := map[string]int{}

	for idx, item := range items {
		label := fmt.Sprintf("item %d (%s/%s/%s/%s)", idx+1, item.Namespace, item.WorkloadKind, item.WorkloadName, item.ContainerName)

		// Required string fields
		if item.Namespace == "" {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: namespace must not be empty", label))
		}
		if item.WorkloadKind == "" {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: workload_kind must not be empty", label))
		}
		if item.WorkloadName == "" {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: workload_name must not be empty", label))
		}
		if item.ContainerName == "" {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: container_name must not be empty", label))
		}

		// Numeric range checks
		if item.CurrentCPURequestMcores < 0 {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: current_cpu_request_mcores must be >= 0", label))
		}
		if item.CurrentMemoryRequestMiB < 0 {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: current_memory_request_mib must be >= 0", label))
		}
		for _, pair := range []struct {
			name string
			val  float64
		}{
			{"cpu_p50_mcores", item.CPUP50Mcores},
			{"cpu_p95_mcores", item.CPUP95Mcores},
			{"cpu_p99_mcores", item.CPUP99Mcores},
			{"cpu_max_mcores", item.CPUMaxMcores},
			{"memory_p50_mib", item.MemoryP50MiB},
			{"memory_p95_mib", item.MemoryP95MiB},
			{"memory_p99_mib", item.MemoryP99MiB},
			{"memory_max_mib", item.MemoryMaxMiB},
		} {
			if pair.val < 0 {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: %s must be >= 0", label, pair.name))
			}
		}
		if item.Samples < 0 {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: samples must be >= 0", label))
		}

		// Duplicate detection
		key := item.Key()
		if prev, exists := seen[key]; exists {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: duplicate entry; first seen at item %d", label, prev+1))
		} else {
			seen[key] = idx
		}

		// Percentile ordering warnings
		if item.CPUP50Mcores > item.CPUP95Mcores {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: cpu_p50 (%.0f) > cpu_p95 (%.0f); metrics may be inconsistent", label, item.CPUP50Mcores, item.CPUP95Mcores))
		}
		if item.CPUP95Mcores > item.CPUP99Mcores {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: cpu_p95 (%.0f) > cpu_p99 (%.0f); metrics may be inconsistent", label, item.CPUP95Mcores, item.CPUP99Mcores))
		}
		if item.CPUP99Mcores > item.CPUMaxMcores {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: cpu_p99 (%.0f) > cpu_max (%.0f); metrics may be inconsistent", label, item.CPUP99Mcores, item.CPUMaxMcores))
		}
		if item.MemoryP50MiB > item.MemoryP95MiB {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: memory_p50 (%.0f) > memory_p95 (%.0f); metrics may be inconsistent", label, item.MemoryP50MiB, item.MemoryP95MiB))
		}
		if item.MemoryP95MiB > item.MemoryP99MiB {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: memory_p95 (%.0f) > memory_p99 (%.0f); metrics may be inconsistent", label, item.MemoryP95MiB, item.MemoryP99MiB))
		}
		if item.MemoryP99MiB > item.MemoryMaxMiB {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: memory_p99 (%.0f) > memory_max (%.0f); metrics may be inconsistent", label, item.MemoryP99MiB, item.MemoryMaxMiB))
		}

		// Low sample warning
		if item.Samples < minSamples {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: only %d samples (minimum is %d); confidence will be low", label, item.Samples, minSamples))
		}

		// Zero usage data warning
		if item.CPUP95Mcores == 0 && item.CPUMaxMcores == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: all CPU usage metrics are zero", label))
		}
		if item.MemoryP95MiB == 0 && item.MemoryMaxMiB == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: all memory usage metrics are zero", label))
		}

		// Zero current requests warning
		if item.CurrentCPURequestMcores == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: current_cpu_request_mcores is 0 (no CPU request set)", label))
		}
		if item.CurrentMemoryRequestMiB == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: current_memory_request_mib is 0 (no memory request set)", label))
		}

		// Missing optional metadata warnings
		if item.Team == "unknown" {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: team is not set", label))
		}
		if item.Environment == "unknown" {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: environment is not set", label))
		}
	}

	return res
}
