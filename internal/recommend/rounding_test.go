package recommend

import "testing"

func TestRoundUpCPU(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 10},
		{5, 10},
		{10, 10},
		{11, 20},
		{100, 100},
		{101, 110},
		{155, 160},
		{220, 220},
	}
	for _, tc := range tests {
		got := RoundUpCPU(tc.input)
		if got != tc.want {
			t.Errorf("RoundUpCPU(%.0f) = %.0f, want %.0f", tc.input, got, tc.want)
		}
	}
}

func TestRoundUpMemory(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 32},
		{16, 32},
		{32, 32},
		{33, 48},
		{256, 256},
		{257, 272},
		{650, 656},
		{1100, 1104},
	}
	for _, tc := range tests {
		got := RoundUpMemory(tc.input)
		if got != tc.want {
			t.Errorf("RoundUpMemory(%.0f) = %.0f, want %.0f", tc.input, got, tc.want)
		}
	}
}
