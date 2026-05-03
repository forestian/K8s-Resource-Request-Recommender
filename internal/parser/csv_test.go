package parser

import (
	"os"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "krr-test-*.csv")
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

const sampleCSV = `cluster,namespace,workload_kind,workload_name,container_name,team,service,environment,current_cpu_request_mcores,current_memory_request_mib,current_cpu_limit_mcores,current_memory_limit_mib,cpu_p50_mcores,cpu_p95_mcores,cpu_p99_mcores,cpu_max_mcores,memory_p50_mib,memory_p95_mib,memory_p99_mib,memory_max_mib,samples,window
prod,default,Deployment,api,api,team-a,api,production,500,1024,1000,2048,80,160,220,300,300,520,650,800,720,7d
`

func TestParseCSV_Success(t *testing.T) {
	path := writeTemp(t, sampleCSV)
	items, err := ParseCSV(path)
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
	if item.CurrentCPURequestMcores != 500 {
		t.Errorf("cpu request: got %v, want 500", item.CurrentCPURequestMcores)
	}
	if item.Samples != 720 {
		t.Errorf("samples: got %v, want 720", item.Samples)
	}
}

func TestParseCSV_DefaultsApplied(t *testing.T) {
	csv := `namespace,workload_kind,workload_name,container_name,current_cpu_request_mcores,current_memory_request_mib,cpu_p50_mcores,cpu_p95_mcores,cpu_p99_mcores,cpu_max_mcores,memory_p50_mib,memory_p95_mib,memory_p99_mib,memory_max_mib,samples
default,Deployment,api,api,100,200,10,20,25,30,50,80,90,100,50
`
	path := writeTemp(t, csv)
	items, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := items[0]
	if item.Cluster != "unknown" {
		t.Errorf("cluster default: got %q, want %q", item.Cluster, "unknown")
	}
	if item.Team != "unknown" {
		t.Errorf("team default: got %q, want %q", item.Team, "unknown")
	}
	if item.Service != "api" {
		t.Errorf("service default: got %q, want %q", item.Service, "api")
	}
	if item.Environment != "unknown" {
		t.Errorf("environment default: got %q, want %q", item.Environment, "unknown")
	}
}

func TestParseCSV_MissingRequiredColumn(t *testing.T) {
	// Missing namespace column
	csv := `workload_kind,workload_name,container_name,current_cpu_request_mcores,current_memory_request_mib,cpu_p50_mcores,cpu_p95_mcores,cpu_p99_mcores,cpu_max_mcores,memory_p50_mib,memory_p95_mib,memory_p99_mib,memory_max_mib,samples
Deployment,api,api,100,200,10,20,25,30,50,80,90,100,50
`
	path := writeTemp(t, csv)
	_, err := ParseCSV(path)
	if err == nil {
		t.Fatal("expected error for missing namespace column")
	}
}
