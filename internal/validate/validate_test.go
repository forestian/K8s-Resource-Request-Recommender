package validate

import (
	"strings"
	"testing"

	"github.com/krr-lite/krr-lite/internal/model"
)

func baseItem() model.UsageItem {
	return model.UsageItem{
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
}

func TestValidate_OK(t *testing.T) {
	result := Validate([]model.UsageItem{baseItem()}, 30)
	if !result.OK() {
		t.Errorf("expected OK, got errors: %v", result.Errors)
	}
}

func TestValidate_MissingNamespace(t *testing.T) {
	item := baseItem()
	item.Namespace = ""
	result := Validate([]model.UsageItem{item}, 30)
	if result.OK() {
		t.Fatal("expected error for missing namespace")
	}
	if !containsError(result.Errors, "namespace must not be empty") {
		t.Errorf("error not found in: %v", result.Errors)
	}
}

func TestValidate_MissingWorkloadName(t *testing.T) {
	item := baseItem()
	item.WorkloadName = ""
	result := Validate([]model.UsageItem{item}, 30)
	if result.OK() {
		t.Fatal("expected error for missing workload_name")
	}
	if !containsError(result.Errors, "workload_name must not be empty") {
		t.Errorf("error not found in: %v", result.Errors)
	}
}

func TestValidate_MissingContainerName(t *testing.T) {
	item := baseItem()
	item.ContainerName = ""
	result := Validate([]model.UsageItem{item}, 30)
	if result.OK() {
		t.Fatal("expected error for missing container_name")
	}
	if !containsError(result.Errors, "container_name must not be empty") {
		t.Errorf("error not found in: %v", result.Errors)
	}
}

func TestValidate_NegativeCPURequest(t *testing.T) {
	item := baseItem()
	item.CurrentCPURequestMcores = -1
	result := Validate([]model.UsageItem{item}, 30)
	if result.OK() {
		t.Fatal("expected error for negative CPU request")
	}
	if !containsError(result.Errors, "current_cpu_request_mcores must be >= 0") {
		t.Errorf("error not found in: %v", result.Errors)
	}
}

func TestValidate_DuplicateItems(t *testing.T) {
	items := []model.UsageItem{baseItem(), baseItem()}
	result := Validate(items, 30)
	if result.OK() {
		t.Fatal("expected error for duplicate items")
	}
	if !containsError(result.Errors, "duplicate entry") {
		t.Errorf("duplicate error not found in: %v", result.Errors)
	}
}

func TestValidate_PercentileOrderingWarning(t *testing.T) {
	item := baseItem()
	item.CPUP50Mcores = 200 // p50 > p95
	result := Validate([]model.UsageItem{item}, 30)
	if !result.OK() {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if !containsWarning(result.Warnings, "cpu_p50") {
		t.Errorf("percentile warning not found in: %v", result.Warnings)
	}
}

func TestValidate_LowSamplesWarning(t *testing.T) {
	item := baseItem()
	item.Samples = 5
	result := Validate([]model.UsageItem{item}, 30)
	if !result.OK() {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if !containsWarning(result.Warnings, "5 samples") {
		t.Errorf("low sample warning not found in: %v", result.Warnings)
	}
}

func containsError(errors []string, substr string) bool {
	for _, e := range errors {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

func containsWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}
