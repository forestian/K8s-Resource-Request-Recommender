package recommend

const (
	StatusMissing        = "missing"
	StatusOverRequested  = "over_requested"
	StatusUnderRequested = "under_requested"
	StatusOK             = "ok"

	ActionSetRequests      = "set_requests"
	ActionReduceRequests   = "reduce_requests"
	ActionIncreaseRequests = "increase_requests"
	ActionReview           = "review"
	ActionNoChange         = "no_change"

	RiskNone   = "none"
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"

	ConfidenceHigh = "high"
	ConfidenceLow  = "low"

	overRequestedThreshold = 0.30 // 30% above recommended
)

// ClassifyCPU returns the CPU status string.
func ClassifyCPU(currentRequest, recommendedRequest, cpuP95 float64) string {
	if currentRequest == 0 {
		return StatusMissing
	}
	if currentRequest > recommendedRequest*(1+overRequestedThreshold) {
		return StatusOverRequested
	}
	if currentRequest < cpuP95 {
		return StatusUnderRequested
	}
	return StatusOK
}

// ClassifyMemory returns the memory status string.
func ClassifyMemory(currentRequest, recommendedRequest, memoryP95 float64) string {
	if currentRequest == 0 {
		return StatusMissing
	}
	if currentRequest > recommendedRequest*(1+overRequestedThreshold) {
		return StatusOverRequested
	}
	if currentRequest < memoryP95 {
		return StatusUnderRequested
	}
	return StatusOK
}

// DetermineAction returns the recommended action given CPU and memory statuses and confidence.
func DetermineAction(cpuStatus, memStatus, confidence string) string {
	if cpuStatus == StatusMissing || memStatus == StatusMissing {
		return ActionSetRequests
	}

	cpuUnder := cpuStatus == StatusUnderRequested
	memUnder := memStatus == StatusUnderRequested
	cpuOver := cpuStatus == StatusOverRequested
	memOver := memStatus == StatusOverRequested

	// Mixed signals
	if (cpuUnder || memUnder) && (cpuOver || memOver) {
		return ActionReview
	}
	if confidence == ConfidenceLow {
		return ActionReview
	}
	if cpuUnder || memUnder {
		return ActionIncreaseRequests
	}
	if cpuOver || memOver {
		return ActionReduceRequests
	}
	return ActionNoChange
}

// riskRank maps a risk string to a numeric rank for comparison.
func riskRank(r string) int {
	switch r {
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	case RiskLow:
		return 1
	default:
		return 0
	}
}

// MaxRisk returns the highest of two risk levels.
func MaxRisk(a, b string) string {
	if riskRank(a) >= riskRank(b) {
		return a
	}
	return b
}

// DetermineRisk computes the overall risk level for a recommendation.
func DetermineRisk(
	cpuStatus, memStatus, confidence, profile string,
	recommendedCPU, recommendedMem, cpuP95, memP99 float64,
) string {
	risk := RiskNone

	switch cpuStatus {
	case StatusMissing:
		risk = MaxRisk(risk, RiskMedium)
	case StatusUnderRequested:
		risk = MaxRisk(risk, RiskMedium)
	case StatusOverRequested:
		risk = MaxRisk(risk, RiskLow)
	}

	switch memStatus {
	case StatusMissing:
		risk = MaxRisk(risk, RiskMedium)
	case StatusUnderRequested:
		risk = MaxRisk(risk, RiskHigh)
	case StatusOverRequested:
		risk = MaxRisk(risk, RiskLow)
	}

	if confidence == ConfidenceLow {
		risk = MaxRisk(risk, RiskMedium)
	}

	// Safety checks: recommendation drops below key percentiles
	if recommendedCPU < cpuP95 {
		risk = MaxRisk(risk, RiskHigh)
	}
	if (profile == "conservative" || profile == "balanced") && recommendedMem < memP99 {
		risk = MaxRisk(risk, RiskHigh)
	}

	return risk
}

// RiskMeetsThreshold returns true when risk is at or above the threshold.
func RiskMeetsThreshold(risk, threshold string) bool {
	if threshold == RiskNone || threshold == "" {
		return false
	}
	return riskRank(risk) >= riskRank(threshold)
}
