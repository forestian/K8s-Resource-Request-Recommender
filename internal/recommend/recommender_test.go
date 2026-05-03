package recommend

import (
	"testing"

	"github.com/krr-lite/krr-lite/internal/model"
)

func makeItem(overrides ...func(*model.UsageItem)) model.UsageItem {
	item := model.UsageItem{
		Cluster:                 "prod",
		Namespace:               "default",
		WorkloadKind:            "Deployment",
		WorkloadName:            "api",
		ContainerName:           "api",
		Team:                    "platform",
		Service:                 "api",
		Environment:             "production",
		CurrentCPURequestMcores: 500,
		CurrentMemoryRequestMiB: 1024,
		CPUP50Mcores:            80,
		CPUP95Mcores:            160,
		CPUP99Mcores:            220,
		CPUMaxMcores:            300,
		MemoryP50MiB:            300,
		MemoryP95MiB:            520,
		MemoryP99MiB:            650,
		MemoryMaxMiB:            800,
		Samples:                 720,
	}
	for _, fn := range overrides {
		fn(&item)
	}
	return item
}

func TestRecommend_BalancedProfile(t *testing.T) {
	// balanced: CPU = max(p95*1.25, p99) = max(160*1.25=200, 220) = 220 → roundup10 = 220
	// balanced: Mem = max(p95*1.20, p99) = max(520*1.20=624, 650) = 650 → roundup16 = 656
	item := makeItem()
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.RecommendedCPURequestMcores != 220 {
		t.Errorf("balanced CPU: got %.0f, want 220", rec.RecommendedCPURequestMcores)
	}
	if rec.RecommendedMemoryRequestMiB != 656 {
		t.Errorf("balanced Memory: got %.0f, want 656", rec.RecommendedMemoryRequestMiB)
	}
}

func TestRecommend_ConservativeProfile(t *testing.T) {
	// conservative: CPU = max(p95*1.5, p99*1.1) = max(160*1.5=240, 220*1.1=242) = 242 → roundup10 = 250
	// conservative: Mem = max(p95*1.3, p99*1.1) = max(520*1.3=676, 650*1.1=715) = 715 → roundup16 = 720
	item := makeItem()
	opts := Options{Profile: "conservative", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.RecommendedCPURequestMcores != 250 {
		t.Errorf("conservative CPU: got %.0f, want 250", rec.RecommendedCPURequestMcores)
	}
	if rec.RecommendedMemoryRequestMiB != 720 {
		t.Errorf("conservative Memory: got %.0f, want 720", rec.RecommendedMemoryRequestMiB)
	}
}

func TestRecommend_AggressiveProfile(t *testing.T) {
	// aggressive: CPU = p95*1.1 = 160*1.1 = 176 → roundup10 = 180
	// aggressive: Mem = p95*1.1 = 520*1.1 = 572 → roundup16 = 576
	item := makeItem()
	opts := Options{Profile: "aggressive", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.RecommendedCPURequestMcores != 180 {
		t.Errorf("aggressive CPU: got %.0f, want 180", rec.RecommendedCPURequestMcores)
	}
	if rec.RecommendedMemoryRequestMiB != 576 {
		t.Errorf("aggressive Memory: got %.0f, want 576", rec.RecommendedMemoryRequestMiB)
	}
}

func TestRecommend_LowSampleConfidence(t *testing.T) {
	item := makeItem(func(i *model.UsageItem) { i.Samples = 5 })
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.Confidence != ConfidenceLow {
		t.Errorf("confidence: got %q, want %q", rec.Confidence, ConfidenceLow)
	}
}

func TestRecommend_MissingCPURequest(t *testing.T) {
	item := makeItem(func(i *model.UsageItem) { i.CurrentCPURequestMcores = 0 })
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.CPUStatus != StatusMissing {
		t.Errorf("cpu status: got %q, want %q", rec.CPUStatus, StatusMissing)
	}
	if rec.Action != ActionSetRequests {
		t.Errorf("action: got %q, want %q", rec.Action, ActionSetRequests)
	}
}

func TestRecommend_OverRequestedClassification(t *testing.T) {
	// balanced: CPU rec = 220, current = 500 → 500 > 220*1.3(=286) → over_requested
	item := makeItem()
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.CPUStatus != StatusOverRequested {
		t.Errorf("cpu status: got %q, want %q", rec.CPUStatus, StatusOverRequested)
	}
}

func TestRecommend_UnderRequestedClassification(t *testing.T) {
	// current CPU (100) < p95 (160), current memory (100) < p95 (520)
	item := makeItem(func(i *model.UsageItem) {
		i.CurrentCPURequestMcores = 100
		i.CurrentMemoryRequestMiB = 100
	})
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend([]model.UsageItem{item}, "test.csv", opts)
	rec := rpt.Recommendations[0]

	if rec.CPUStatus != StatusUnderRequested {
		t.Errorf("cpu status: got %q, want %q", rec.CPUStatus, StatusUnderRequested)
	}
	if rec.Action != ActionIncreaseRequests {
		t.Errorf("action: got %q, want %q", rec.Action, ActionIncreaseRequests)
	}
}

func TestRecommend_SummaryAccumulation(t *testing.T) {
	items := []model.UsageItem{
		makeItem(), // over_requested CPU → reduce
		makeItem(func(i *model.UsageItem) {
			i.WorkloadName = "other"
			i.ContainerName = "other"
			i.CurrentCPURequestMcores = 100 // under_requested CPU
			i.CurrentMemoryRequestMiB = 100 // under_requested memory too (avoid mixed signals)
		}),
	}
	opts := Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := Recommend(items, "test.csv", opts)

	if rpt.Summary.ReduceRequests != 1 {
		t.Errorf("ReduceRequests: got %d, want 1", rpt.Summary.ReduceRequests)
	}
	if rpt.Summary.IncreaseRequests != 1 {
		t.Errorf("IncreaseRequests: got %d, want 1", rpt.Summary.IncreaseRequests)
	}
}

func TestRiskMeetsThreshold(t *testing.T) {
	tests := []struct {
		risk      string
		threshold string
		want      bool
	}{
		{"high", "high", true},
		{"high", "medium", true},
		{"high", "low", true},
		{"medium", "high", false},
		{"medium", "medium", true},
		{"medium", "low", true},
		{"low", "high", false},
		{"low", "medium", false},
		{"low", "low", true},
		{"none", "high", false},
		{"low", "none", false},
	}
	for _, tc := range tests {
		got := RiskMeetsThreshold(tc.risk, tc.threshold)
		if got != tc.want {
			t.Errorf("RiskMeetsThreshold(%q, %q) = %v, want %v", tc.risk, tc.threshold, got, tc.want)
		}
	}
}
