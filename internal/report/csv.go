package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

var csvHeader = []string{
	"cluster", "namespace", "workload_kind", "workload_name", "container_name",
	"team", "service", "environment",
	"current_cpu_request_mcores", "recommended_cpu_request_mcores",
	"current_memory_request_mib", "recommended_memory_request_mib",
	"cpu_status", "memory_status", "action", "risk", "confidence",
	"cpu_request_delta_mcores", "memory_request_delta_mib",
	"cpu_reduction_percent", "memory_reduction_percent",
	"reasons", "suggestions",
}

// WriteCSV writes a CSV report to w.
func WriteCSV(w io.Writer, report *model.Report, includeUnchanged bool) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return fmt.Errorf("cannot write CSV header: %w", err)
	}

	for _, rec := range report.Recommendations {
		if !includeUnchanged && rec.Action == "no_change" {
			continue
		}
		row := []string{
			rec.Cluster,
			rec.Namespace,
			rec.WorkloadKind,
			rec.WorkloadName,
			rec.ContainerName,
			rec.Team,
			rec.Service,
			rec.Environment,
			fmt.Sprintf("%.2f", rec.CurrentCPURequestMcores),
			fmt.Sprintf("%.2f", rec.RecommendedCPURequestMcores),
			fmt.Sprintf("%.2f", rec.CurrentMemoryRequestMiB),
			fmt.Sprintf("%.2f", rec.RecommendedMemoryRequestMiB),
			rec.CPUStatus,
			rec.MemoryStatus,
			rec.Action,
			rec.Risk,
			rec.Confidence,
			fmt.Sprintf("%.2f", rec.CPURequestDeltaMcores),
			fmt.Sprintf("%.2f", rec.MemoryRequestDeltaMiB),
			fmt.Sprintf("%.2f", rec.CPUReductionPercent),
			fmt.Sprintf("%.2f", rec.MemoryReductionPercent),
			strings.Join(rec.Reasons, "; "),
			strings.Join(rec.Suggestions, "; "),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("cannot write CSV row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
