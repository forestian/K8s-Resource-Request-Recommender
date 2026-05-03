package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/krr-lite/krr-lite/internal/model"
)

func makeTestReport() *model.Report {
	return &model.Report{
		Profile:    "balanced",
		InputFile:  "usage.csv",
		TotalItems: 2,
		Summary: model.Summary{
			ReduceRequests:              1,
			IncreaseRequests:            1,
			LowRisk:                     1,
			HighRisk:                    1,
			PotentialCPUReductionMcores: 280,
			PotentialMemoryReductionMiB: 368,
		},
		Recommendations: []model.Recommendation{
			{
				Cluster:                     "prod",
				Namespace:                   "default",
				WorkloadKind:                "Deployment",
				WorkloadName:                "api",
				ContainerName:               "api",
				Team:                        "platform",
				Service:                     "api",
				Environment:                 "production",
				CurrentCPURequestMcores:     500,
				CurrentMemoryRequestMiB:     1024,
				RecommendedCPURequestMcores: 220,
				RecommendedMemoryRequestMiB: 656,
				CPUStatus:                   "over_requested",
				MemoryStatus:                "over_requested",
				Action:                      "reduce_requests",
				Risk:                        "low",
				Confidence:                  "high",
				CPURequestDeltaMcores:       280,
				MemoryRequestDeltaMiB:       368,
				CPUReductionPercent:         56,
				MemoryReductionPercent:      35.9,
				Reasons:                     []string{"Current CPU request is more than 30% above recommended request."},
				Suggestions:                 []string{"Review recent usage before applying."},
			},
			{
				Namespace:                   "monitoring",
				WorkloadName:                "loki",
				ContainerName:               "loki",
				Action:                      "no_change",
				Risk:                        "none",
				Confidence:                  "high",
				CurrentCPURequestMcores:     2000,
				RecommendedCPURequestMcores: 2000,
				CPUStatus:                   "ok",
				MemoryStatus:                "ok",
			},
		},
		Warnings: []string{"test warning"},
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	rpt := makeTestReport()
	if err := WriteText(&buf, rpt, false); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "K8s Resource Request Recommender") {
		t.Error("missing title")
	}
	if !strings.Contains(out, "reduce_requests") {
		t.Error("missing action")
	}
	// no_change should be omitted
	if strings.Contains(out, "no_change") {
		t.Error("no_change should be omitted when include-unchanged=false")
	}
}

func TestWriteText_IncludeUnchanged(t *testing.T) {
	var buf bytes.Buffer
	rpt := makeTestReport()
	if err := WriteText(&buf, rpt, true); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "no_change") {
		t.Error("no_change should be present when include-unchanged=true")
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	rpt := makeTestReport()
	if err := WriteJSON(&buf, rpt); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	var decoded model.Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}
	if decoded.Profile != "balanced" {
		t.Errorf("profile: got %q, want balanced", decoded.Profile)
	}
	// JSON always includes all recommendations
	if len(decoded.Recommendations) != 2 {
		t.Errorf("JSON should include all 2 recommendations, got %d", len(decoded.Recommendations))
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	rpt := makeTestReport()
	if err := WriteMarkdown(&buf, rpt, false); err != nil {
		t.Fatalf("WriteMarkdown error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# K8s Resource Request Recommender") {
		t.Error("missing markdown title")
	}
	if !strings.Contains(out, "| Set requests |") {
		t.Error("missing summary table")
	}
	if !strings.Contains(out, "reduce_requests") {
		t.Error("missing recommendation")
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	rpt := makeTestReport()
	if err := WriteCSV(&buf, rpt, false); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// header + 1 data row (no_change omitted)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if !strings.Contains(lines[0], "cluster") {
		t.Error("header missing 'cluster'")
	}
	if !strings.Contains(lines[1], "over_requested") {
		t.Error("data row missing 'over_requested'")
	}
}
