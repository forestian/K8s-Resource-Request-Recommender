package recommend

import (
	"fmt"
	"math"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

// Options configures the recommender.
type Options struct {
	Profile    string
	MinSamples int
}

// Recommend generates a Report from usage items.
func Recommend(items []model.UsageItem, inputFile string, opts Options) (*model.Report, []string) {
	profile, ok := Profiles[opts.Profile]
	if !ok {
		profile = Profiles["balanced"]
	}

	var warnings []string
	var recs []model.Recommendation
	summary := model.Summary{}

	for _, item := range items {
		rec, itemWarnings := computeRecommendation(item, profile, opts.MinSamples)
		warnings = append(warnings, itemWarnings...)

		// Accumulate summary counts
		switch rec.Action {
		case ActionSetRequests:
			summary.SetRequests++
		case ActionReduceRequests:
			summary.ReduceRequests++
		case ActionIncreaseRequests:
			summary.IncreaseRequests++
		case ActionReview:
			summary.Review++
		case ActionNoChange:
			summary.NoChange++
		}

		switch rec.Risk {
		case RiskHigh:
			summary.HighRisk++
		case RiskMedium:
			summary.MediumRisk++
		case RiskLow:
			summary.LowRisk++
		}

		if rec.CPURequestDeltaMcores > 0 {
			summary.PotentialCPUReductionMcores += rec.CPURequestDeltaMcores
		}
		if rec.MemoryRequestDeltaMiB > 0 {
			summary.PotentialMemoryReductionMiB += rec.MemoryRequestDeltaMiB
		}

		recs = append(recs, rec)
	}

	return &model.Report{
		Profile:         opts.Profile,
		InputFile:       inputFile,
		TotalItems:      len(items),
		Summary:         summary,
		Recommendations: recs,
		Warnings:        warnings,
	}, warnings
}

func computeRecommendation(item model.UsageItem, profile Profile, minSamples int) (model.Recommendation, []string) {
	var warnings []string

	// Determine confidence
	confidence := ConfidenceHigh
	if item.Samples < minSamples {
		confidence = ConfidenceLow
		warnings = append(warnings, fmt.Sprintf(
			"%s/%s/%s: only %d samples (minimum %d); confidence is low",
			item.Namespace, item.WorkloadName, item.ContainerName, item.Samples, minSamples,
		))
	}

	// Compute raw recommended values from profile
	rawCPU := computeRawCPU(item, profile)
	rawMem := computeRawMemory(item, profile)

	// Apply minimums and rounding
	recCPU := RoundUpCPU(rawCPU)
	recMem := RoundUpMemory(rawMem)

	// Classify
	cpuStatus := ClassifyCPU(item.CurrentCPURequestMcores, recCPU, item.CPUP95Mcores)
	memStatus := ClassifyMemory(item.CurrentMemoryRequestMiB, recMem, item.MemoryP95MiB)
	action := DetermineAction(cpuStatus, memStatus, confidence)
	risk := DetermineRisk(cpuStatus, memStatus, confidence, profile.Name, recCPU, recMem, item.CPUP95Mcores, item.MemoryP99MiB)

	// Deltas
	cpuDelta := item.CurrentCPURequestMcores - recCPU
	memDelta := item.CurrentMemoryRequestMiB - recMem

	cpuReductionPct := 0.0
	if item.CurrentCPURequestMcores > 0 {
		cpuReductionPct = math.Round((cpuDelta/item.CurrentCPURequestMcores)*10000) / 100
	}
	memReductionPct := 0.0
	if item.CurrentMemoryRequestMiB > 0 {
		memReductionPct = math.Round((memDelta/item.CurrentMemoryRequestMiB)*10000) / 100
	}

	// Generate reasons and suggestions
	reasons, suggestions := generateReasonsSuggestions(item, cpuStatus, memStatus, action, profile.Name, confidence, recCPU, recMem)

	// Emit additional warnings
	if profile.Name == "aggressive" && recMem < item.MemoryP99MiB {
		warnings = append(warnings, fmt.Sprintf(
			"%s/%s/%s: aggressive profile recommends memory (%.0f MiB) below p99 (%.0f MiB); OOM risk",
			item.Namespace, item.WorkloadName, item.ContainerName, recMem, item.MemoryP99MiB,
		))
	}
	if cpuReductionPct > 70 {
		warnings = append(warnings, fmt.Sprintf(
			"%s/%s/%s: CPU reduction is %.0f%%; verify usage data before applying",
			item.Namespace, item.WorkloadName, item.ContainerName, cpuReductionPct,
		))
	}
	if memReductionPct > 70 {
		warnings = append(warnings, fmt.Sprintf(
			"%s/%s/%s: memory reduction is %.0f%%; verify usage data before applying",
			item.Namespace, item.WorkloadName, item.ContainerName, memReductionPct,
		))
	}

	rec := model.Recommendation{
		Cluster:                     item.Cluster,
		Namespace:                   item.Namespace,
		WorkloadKind:                item.WorkloadKind,
		WorkloadName:                item.WorkloadName,
		ContainerName:               item.ContainerName,
		Team:                        item.Team,
		Service:                     item.Service,
		Environment:                 item.Environment,
		CurrentCPURequestMcores:     item.CurrentCPURequestMcores,
		CurrentMemoryRequestMiB:     item.CurrentMemoryRequestMiB,
		RecommendedCPURequestMcores: recCPU,
		RecommendedMemoryRequestMiB: recMem,
		CPUStatus:                   cpuStatus,
		MemoryStatus:                memStatus,
		Action:                      action,
		Risk:                        risk,
		Confidence:                  confidence,
		CPURequestDeltaMcores:       cpuDelta,
		MemoryRequestDeltaMiB:       memDelta,
		CPUReductionPercent:         cpuReductionPct,
		MemoryReductionPercent:      memReductionPct,
		Reasons:                     reasons,
		Suggestions:                 suggestions,
	}
	return rec, warnings
}

func computeRawCPU(item model.UsageItem, profile Profile) float64 {
	p95val := item.CPUP95Mcores * profile.CPUP95Mult
	if !profile.UseP99ForCPU {
		return p95val
	}
	p99val := item.CPUP99Mcores * profile.CPUP99Mult
	return math.Max(p95val, p99val)
}

func computeRawMemory(item model.UsageItem, profile Profile) float64 {
	p95val := item.MemoryP95MiB * profile.MemoryP95Mult
	if !profile.UseP99ForMemory {
		return p95val
	}
	p99val := item.MemoryP99MiB * profile.MemoryP99Mult
	return math.Max(p95val, p99val)
}

func generateReasonsSuggestions(
	item model.UsageItem,
	cpuStatus, memStatus, action, profileName, confidence string,
	recCPU, recMem float64,
) (reasons, suggestions []string) {
	switch cpuStatus {
	case StatusMissing:
		reasons = append(reasons, "CPU request is not set.")
	case StatusOverRequested:
		reasons = append(reasons, "Current CPU request is more than 30% above recommended request.")
	case StatusUnderRequested:
		reasons = append(reasons, "Current CPU request is below observed p95 CPU usage.")
	}

	switch memStatus {
	case StatusMissing:
		reasons = append(reasons, "Memory request is not set.")
	case StatusOverRequested:
		reasons = append(reasons, "Current memory request is more than 30% above recommended request.")
	case StatusUnderRequested:
		reasons = append(reasons, "Current memory request is below observed p95 memory usage.")
	}

	if confidence == ConfidenceLow {
		reasons = append(reasons, fmt.Sprintf("Sample count (%d) is below minimum threshold; recommendation confidence is low.", item.Samples))
	}

	if profileName == "aggressive" && recMem < item.MemoryP99MiB {
		reasons = append(reasons, fmt.Sprintf("Aggressive profile recommends memory (%.0f MiB) below p99 (%.0f MiB).", recMem, item.MemoryP99MiB))
	}

	// Suggestions based on action
	switch action {
	case ActionSetRequests:
		suggestions = append(suggestions, "Set explicit CPU and memory requests to improve scheduler accuracy.")
		suggestions = append(suggestions, "Apply gradually and monitor throttling, OOM kills, and latency.")
	case ActionReduceRequests:
		suggestions = append(suggestions, "Review recent usage before applying.")
		suggestions = append(suggestions, "Consider updating Helm values or manifests with the recommended requests.")
		suggestions = append(suggestions, "Apply gradually and monitor throttling, OOM kills, and latency.")
	case ActionIncreaseRequests:
		suggestions = append(suggestions, "Increase requests to avoid throttling or OOM kills.")
		suggestions = append(suggestions, "Apply gradually and monitor OOMKills after applying.")
	case ActionReview:
		suggestions = append(suggestions, "Investigate mixed or conflicting usage signals before making changes.")
		suggestions = append(suggestions, "Review recent usage data for correctness.")
	}

	if memStatus == StatusUnderRequested {
		suggestions = append(suggestions, "Review for potential OOM kill risk.")
	}

	// De-duplicate suggestions
	suggestions = deduplicate(suggestions)
	return reasons, suggestions
}

func deduplicate(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[strings.TrimSpace(s)] {
			seen[strings.TrimSpace(s)] = true
			out = append(out, s)
		}
	}
	return out
}
