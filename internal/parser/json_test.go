package parser

import (
	"os"
	"testing"
)

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "krr-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

const sampleJSON = `{
  "items": [
    {
      "cluster": "prod",
      "namespace": "default",
      "workload_kind": "Deployment",
      "workload_name": "api",
      "container_name": "api",
      "team": "platform",
      "service": "api",
      "environment": "production",
      "current_cpu_request_mcores": 500,
      "current_memory_request_mib": 1024,
      "cpu_p50_mcores": 80,
      "cpu_p95_mcores": 160,
      "cpu_p99_mcores": 220,
      "cpu_max_mcores": 300,
      "memory_p50_mib": 300,
      "memory_p95_mib": 520,
      "memory_p99_mib": 650,
      "memory_max_mib": 800,
      "samples": 720,
      "window": "7d"
    }
  ]
}`

func TestParseJSON_Success(t *testing.T) {
	path := writeTempJSON(t, sampleJSON)
	items, err := ParseJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Namespace != "default" {
		t.Errorf("namespace: got %q, want %q", item.Namespace, "default")
	}
	if item.CPUP95Mcores != 160 {
		t.Errorf("cpu_p95: got %v, want 160", item.CPUP95Mcores)
	}
}

func TestParseJSON_DefaultsApplied(t *testing.T) {
	j := `{"items":[{"namespace":"default","workload_kind":"Deployment","workload_name":"api","container_name":"api","current_cpu_request_mcores":100,"current_memory_request_mib":200,"cpu_p50_mcores":10,"cpu_p95_mcores":20,"cpu_p99_mcores":25,"cpu_max_mcores":30,"memory_p50_mib":50,"memory_p95_mib":80,"memory_p99_mib":90,"memory_max_mib":100,"samples":50}]}`
	path := writeTempJSON(t, j)
	items, err := ParseJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := items[0]
	if item.Cluster != "unknown" {
		t.Errorf("cluster default: got %q", item.Cluster)
	}
	if item.Service != "api" {
		t.Errorf("service default: got %q, want %q", item.Service, "api")
	}
}

func TestParseJSON_Invalid(t *testing.T) {
	path := writeTempJSON(t, `not json`)
	_, err := ParseJSON(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
