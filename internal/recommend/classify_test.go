package recommend

import "testing"

func TestClassifyCPU_Missing(t *testing.T) {
	status := ClassifyCPU(0, 200, 160)
	if status != StatusMissing {
		t.Errorf("got %q, want %q", status, StatusMissing)
	}
}

func TestClassifyCPU_OverRequested(t *testing.T) {
	// current (500) > recommended (200) * 1.3 (= 260)
	status := ClassifyCPU(500, 200, 160)
	if status != StatusOverRequested {
		t.Errorf("got %q, want %q", status, StatusOverRequested)
	}
}

func TestClassifyCPU_UnderRequested(t *testing.T) {
	// current (100) < p95 (160)
	status := ClassifyCPU(100, 200, 160)
	if status != StatusUnderRequested {
		t.Errorf("got %q, want %q", status, StatusUnderRequested)
	}
}

func TestClassifyCPU_OK(t *testing.T) {
	// current (220) >= p95 (160), and current (220) <= recommended (200) * 1.3 (260)
	status := ClassifyCPU(220, 200, 160)
	if status != StatusOK {
		t.Errorf("got %q, want %q", status, StatusOK)
	}
}

func TestClassifyMemory_Missing(t *testing.T) {
	status := ClassifyMemory(0, 650, 520)
	if status != StatusMissing {
		t.Errorf("got %q, want %q", status, StatusMissing)
	}
}

func TestDetermineAction_SetRequests(t *testing.T) {
	action := DetermineAction(StatusMissing, StatusOK, ConfidenceHigh)
	if action != ActionSetRequests {
		t.Errorf("got %q, want %q", action, ActionSetRequests)
	}
}

func TestDetermineAction_ReduceRequests(t *testing.T) {
	action := DetermineAction(StatusOverRequested, StatusOverRequested, ConfidenceHigh)
	if action != ActionReduceRequests {
		t.Errorf("got %q, want %q", action, ActionReduceRequests)
	}
}

func TestDetermineAction_IncreaseRequests(t *testing.T) {
	action := DetermineAction(StatusUnderRequested, StatusOK, ConfidenceHigh)
	if action != ActionIncreaseRequests {
		t.Errorf("got %q, want %q", action, ActionIncreaseRequests)
	}
}

func TestDetermineAction_Review_LowConfidence(t *testing.T) {
	action := DetermineAction(StatusOK, StatusOK, ConfidenceLow)
	if action != ActionReview {
		t.Errorf("got %q, want %q", action, ActionReview)
	}
}

func TestDetermineAction_Review_Mixed(t *testing.T) {
	action := DetermineAction(StatusOverRequested, StatusUnderRequested, ConfidenceHigh)
	if action != ActionReview {
		t.Errorf("got %q, want %q", action, ActionReview)
	}
}

func TestDetermineAction_NoChange(t *testing.T) {
	action := DetermineAction(StatusOK, StatusOK, ConfidenceHigh)
	if action != ActionNoChange {
		t.Errorf("got %q, want %q", action, ActionNoChange)
	}
}
