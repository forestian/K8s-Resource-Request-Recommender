package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

// WriteMarkdown writes a GitHub-flavored markdown report to w.
func WriteMarkdown(w io.Writer, report *model.Report, includeUnchanged bool) error {
	fmt.Fprintln(w, "# K8s Resource Request Recommender")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "_Profile: **%s** | Input: `%s` | Items analyzed: **%d**_\n", report.Profile, report.InputFile, report.TotalItems)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Summary")
	fmt.Fprintln(w)

	s := report.Summary
	fmt.Fprintln(w, "| Metric | Count |")
	fmt.Fprintln(w, "|---|---:|")
	fmt.Fprintf(w, "| Set requests | %d |\n", s.SetRequests)
	fmt.Fprintf(w, "| Reduce requests | %d |\n", s.ReduceRequests)
	fmt.Fprintf(w, "| Increase requests | %d |\n", s.IncreaseRequests)
	fmt.Fprintf(w, "| Review | %d |\n", s.Review)
	fmt.Fprintf(w, "| No change | %d |\n", s.NoChange)
	fmt.Fprintf(w, "| High risk | %d |\n", s.HighRisk)
	fmt.Fprintf(w, "| Medium risk | %d |\n", s.MediumRisk)
	fmt.Fprintf(w, "| Low risk | %d |\n", s.LowRisk)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "| Potential CPU reduction | `%.0fm` |\n", s.PotentialCPUReductionMcores)
	fmt.Fprintf(w, "| Potential memory reduction | `%.0fMi` |\n", s.PotentialMemoryReductionMiB)
	fmt.Fprintln(w)

	if len(report.Warnings) > 0 {
		fmt.Fprintln(w, "## Warnings")
		fmt.Fprintln(w)
		for _, warning := range report.Warnings {
			fmt.Fprintf(w, "- %s\n", warning)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "## Recommendations")
	fmt.Fprintln(w)

	printed := 0
	for _, rec := range report.Recommendations {
		if !includeUnchanged && rec.Action == "no_change" {
			continue
		}
		if err := writeMarkdownRec(w, rec); err != nil {
			return err
		}
		printed++
	}
	if printed == 0 {
		fmt.Fprintln(w, "_No recommendations to display._")
	}

	return nil
}

func writeMarkdownRec(w io.Writer, rec model.Recommendation) error {
	fmt.Fprintf(w, "### %s / %s / %s/%s / %s\n",
		strings.ToUpper(rec.Risk),
		rec.Action,
		rec.Namespace,
		rec.WorkloadName,
		rec.ContainerName,
	)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "| Field | Current | Recommended |")
	fmt.Fprintln(w, "|---|---:|---:|")
	fmt.Fprintf(w, "| CPU request | `%s` | `%s` |\n",
		fmtCPU(rec.CurrentCPURequestMcores),
		fmtCPU(rec.RecommendedCPURequestMcores),
	)
	fmt.Fprintf(w, "| Memory request | `%s` | `%s` |\n",
		fmtMem(rec.CurrentMemoryRequestMiB),
		fmtMem(rec.RecommendedMemoryRequestMiB),
	)
	fmt.Fprintf(w, "| CPU status | `%s` | — |\n", rec.CPUStatus)
	fmt.Fprintf(w, "| Memory status | `%s` | — |\n", rec.MemoryStatus)
	fmt.Fprintf(w, "| Confidence | `%s` | — |\n", rec.Confidence)
	fmt.Fprintln(w)

	if len(rec.Reasons) > 0 {
		fmt.Fprintln(w, "**Reasons:**")
		fmt.Fprintln(w)
		for _, r := range rec.Reasons {
			fmt.Fprintf(w, "- %s\n", r)
		}
		fmt.Fprintln(w)
	}

	if len(rec.Suggestions) > 0 {
		fmt.Fprintln(w, "**Suggestions:**")
		fmt.Fprintln(w)
		for _, s := range rec.Suggestions {
			fmt.Fprintf(w, "- %s\n", s)
		}
		fmt.Fprintln(w)
	}

	return nil
}
