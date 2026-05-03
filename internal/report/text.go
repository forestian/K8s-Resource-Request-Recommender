package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

// WriteText writes a human-readable text report to w.
func WriteText(w io.Writer, report *model.Report, includeUnchanged bool) error {
	fmt.Fprintln(w, "K8s Resource Request Recommender")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Profile:        %s\n", report.Profile)
	fmt.Fprintf(w, "Input:          %s\n", report.InputFile)
	fmt.Fprintf(w, "Items analyzed: %d\n", report.TotalItems)
	fmt.Fprintln(w)

	s := report.Summary
	fmt.Fprintln(w, "Summary:")
	fmt.Fprintf(w, "  Set requests:     %d\n", s.SetRequests)
	fmt.Fprintf(w, "  Reduce requests:  %d\n", s.ReduceRequests)
	fmt.Fprintf(w, "  Increase requests: %d\n", s.IncreaseRequests)
	fmt.Fprintf(w, "  Review:           %d\n", s.Review)
	fmt.Fprintf(w, "  No change:        %d\n", s.NoChange)
	fmt.Fprintf(w, "  High risk:        %d\n", s.HighRisk)
	fmt.Fprintf(w, "  Medium risk:      %d\n", s.MediumRisk)
	fmt.Fprintf(w, "  Low risk:         %d\n", s.LowRisk)
	fmt.Fprintf(w, "  Potential CPU request reduction:    %.0fm\n", s.PotentialCPUReductionMcores)
	fmt.Fprintf(w, "  Potential memory request reduction: %.0fMi\n", s.PotentialMemoryReductionMiB)
	fmt.Fprintln(w)

	if len(report.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings:")
		for _, w2 := range report.Warnings {
			fmt.Fprintf(w, "  - %s\n", w2)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Recommendations:")
	fmt.Fprintln(w)

	printed := 0
	for _, rec := range report.Recommendations {
		if !includeUnchanged && rec.Action == "no_change" {
			continue
		}
		if err := writeTextRec(w, rec); err != nil {
			return err
		}
		printed++
	}
	if printed == 0 {
		fmt.Fprintln(w, "  No recommendations to display.")
	}

	return nil
}

func writeTextRec(w io.Writer, rec model.Recommendation) error {
	fmt.Fprintf(w, "[%s] %s %s/%s container=%s\n",
		strings.ToUpper(rec.Risk),
		rec.Action,
		rec.Namespace,
		rec.WorkloadName,
		rec.ContainerName,
	)
	fmt.Fprintf(w, "CPU request:    %s -> %s\n",
		fmtCPU(rec.CurrentCPURequestMcores),
		fmtCPU(rec.RecommendedCPURequestMcores),
	)
	fmt.Fprintf(w, "Memory request: %s -> %s\n",
		fmtMem(rec.CurrentMemoryRequestMiB),
		fmtMem(rec.RecommendedMemoryRequestMiB),
	)
	fmt.Fprintf(w, "CPU status:     %s\n", rec.CPUStatus)
	fmt.Fprintf(w, "Memory status:  %s\n", rec.MemoryStatus)
	fmt.Fprintf(w, "Confidence:     %s\n", rec.Confidence)

	if len(rec.Reasons) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Reasons:")
		for _, r := range rec.Reasons {
			fmt.Fprintf(w, "  - %s\n", r)
		}
	}

	if len(rec.Suggestions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Suggestions:")
		for _, s := range rec.Suggestions {
			fmt.Fprintf(w, "  - %s\n", s)
		}
	}
	fmt.Fprintln(w)
	return nil
}

// fmtCPU formats a millicores value as "Xm".
func fmtCPU(mcores float64) string {
	return fmt.Sprintf("%.0fm", mcores)
}

// fmtMem formats a MiB value as "XMi".
func fmtMem(mib float64) string {
	return fmt.Sprintf("%.0fMi", mib)
}
