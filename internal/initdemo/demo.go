package initdemo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krr-lite/krr-lite/internal/model"
	"github.com/krr-lite/krr-lite/internal/recommend"
	"github.com/krr-lite/krr-lite/internal/report"
)

// SampleCSV is the example usage.csv content.
const SampleCSV = `cluster,namespace,workload_kind,workload_name,container_name,team,service,environment,current_cpu_request_mcores,current_memory_request_mib,current_cpu_limit_mcores,current_memory_limit_mib,cpu_p50_mcores,cpu_p95_mcores,cpu_p99_mcores,cpu_max_mcores,memory_p50_mib,memory_p95_mib,memory_p99_mib,memory_max_mib,samples,window
prod-kr,default,Deployment,api,api,platform,api,production,500,1024,1000,2048,80,160,220,300,300,520,650,800,720,7d
prod-kr,default,Deployment,worker,worker,platform,worker,production,1000,2048,2000,4096,120,250,320,450,500,900,1100,1300,720,7d
prod-kr,monitoring,StatefulSet,loki,loki,platform,observability,production,2000,4096,4000,8192,900,1800,2200,2600,2500,3800,4500,5200,720,7d
`

// SampleJSON is the example usage.json content.
const SampleJSON = `{
  "items": [
    {
      "cluster": "prod-kr",
      "namespace": "default",
      "workload_kind": "Deployment",
      "workload_name": "api",
      "container_name": "api",
      "team": "platform",
      "service": "api",
      "environment": "production",
      "current_cpu_request_mcores": 500,
      "current_memory_request_mib": 1024,
      "current_cpu_limit_mcores": 1000,
      "current_memory_limit_mib": 2048,
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
    },
    {
      "cluster": "prod-kr",
      "namespace": "default",
      "workload_kind": "Deployment",
      "workload_name": "worker",
      "container_name": "worker",
      "team": "platform",
      "service": "worker",
      "environment": "production",
      "current_cpu_request_mcores": 1000,
      "current_memory_request_mib": 2048,
      "current_cpu_limit_mcores": 2000,
      "current_memory_limit_mib": 4096,
      "cpu_p50_mcores": 120,
      "cpu_p95_mcores": 250,
      "cpu_p99_mcores": 320,
      "cpu_max_mcores": 450,
      "memory_p50_mib": 500,
      "memory_p95_mib": 900,
      "memory_p99_mib": 1100,
      "memory_max_mib": 1300,
      "samples": 720,
      "window": "7d"
    },
    {
      "cluster": "prod-kr",
      "namespace": "monitoring",
      "workload_kind": "StatefulSet",
      "workload_name": "loki",
      "container_name": "loki",
      "team": "platform",
      "service": "observability",
      "environment": "production",
      "current_cpu_request_mcores": 2000,
      "current_memory_request_mib": 4096,
      "current_cpu_limit_mcores": 4000,
      "current_memory_limit_mib": 8192,
      "cpu_p50_mcores": 900,
      "cpu_p95_mcores": 1800,
      "cpu_p99_mcores": 2200,
      "cpu_max_mcores": 2600,
      "memory_p50_mib": 2500,
      "memory_p95_mib": 3800,
      "memory_p99_mib": 4500,
      "memory_max_mib": 5200,
      "samples": 720,
      "window": "7d"
    }
  ]
}
`

const demoReadme = `# krr-lite Demo

This directory contains example usage data and generated recommendation reports.

## Files

| File | Description |
|---|---|
| usage.csv | Example usage data in CSV format |
| usage.json | Example usage data in JSON format |
| reports/recommendations.md | Recommendations in Markdown (suitable for GitHub PR comments) |
| reports/recommendations.json | Recommendations in JSON format |
| reports/recommendations.csv | Recommendations in CSV format |

## How to use

### Edit the usage data

Edit ` + "`usage.csv`" + ` or ` + "`usage.json`" + ` with data exported from your monitoring system
(Prometheus, Mimir, Kubecost, or a custom script).

### Validate your input

` + "```" + `sh
krr-lite validate --file usage.csv
krr-lite validate --file usage.json
` + "```" + `

### Generate recommendations

` + "```" + `sh
# Default (balanced profile, text output)
krr-lite recommend --file usage.csv

# Conservative profile
krr-lite recommend --file usage.csv --profile conservative

# Aggressive profile
krr-lite recommend --file usage.csv --profile aggressive

# Markdown output (good for PR comments)
krr-lite recommend --file usage.csv --format markdown --output reports/recommendations.md --force

# JSON output
krr-lite recommend --file usage.csv --format json --output reports/recommendations.json --force

# CSV output
krr-lite recommend --file usage.csv --format csv --output reports/recommendations.csv --force
` + "```" + `

## Recommendation profiles

| Profile | Description |
|---|---|
| conservative | Higher safety margins; use when risk tolerance is low |
| balanced | Default; good balance between safety and cost reduction |
| aggressive | Stronger cost reduction; accept more risk |

## Important

**This tool does not apply changes automatically.**
Always review recommendations before updating manifests, Helm values, or Kustomize overlays.
Monitor for throttling and OOM kills after applying changes.
`

// Generate creates the demo directory with sample files and reports.
func Generate(outputDir string, force bool) error {
	// Check if output dir exists
	if _, err := os.Stat(outputDir); err == nil {
		if !force {
			return fmt.Errorf("output directory %q already exists; use --force to overwrite", outputDir)
		}
	}

	reportsDir := filepath.Join(outputDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Write usage files
	if err := writeFile(filepath.Join(outputDir, "usage.csv"), SampleCSV, force); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(outputDir, "usage.json"), SampleJSON, force); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(outputDir, "README.md"), demoReadme, force); err != nil {
		return err
	}

	// Generate reports from sample data
	items, err := parseSampleCSV()
	if err != nil {
		return fmt.Errorf("internal error parsing sample data: %w", err)
	}

	opts := recommend.Options{Profile: "balanced", MinSamples: 30}
	rpt, _ := recommend.Recommend(items, "usage.csv", opts)

	// Markdown report
	mdPath := filepath.Join(reportsDir, "recommendations.md")
	if err := writeReport(mdPath, "markdown", rpt, force); err != nil {
		return err
	}

	// JSON report
	jsonPath := filepath.Join(reportsDir, "recommendations.json")
	if err := writeReport(jsonPath, "json", rpt, force); err != nil {
		return err
	}

	// CSV report
	csvPath := filepath.Join(reportsDir, "recommendations.csv")
	if err := writeReport(csvPath, "csv", rpt, force); err != nil {
		return err
	}

	fmt.Printf("Created %s\n", outputDir)
	fmt.Printf("  %s\n", filepath.Join(outputDir, "README.md"))
	fmt.Printf("  %s\n", filepath.Join(outputDir, "usage.csv"))
	fmt.Printf("  %s\n", filepath.Join(outputDir, "usage.json"))
	fmt.Printf("  %s\n", mdPath)
	fmt.Printf("  %s\n", jsonPath)
	fmt.Printf("  %s\n", csvPath)

	return nil
}

func writeFile(path, content string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file %q already exists; use --force to overwrite", path)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func writeReport(path, format string, rpt *model.Report, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file %q already exists; use --force to overwrite", path)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create %q: %w", path, err)
	}
	defer f.Close()

	switch format {
	case "markdown":
		return report.WriteMarkdown(f, rpt, false)
	case "json":
		return report.WriteJSON(f, rpt)
	case "csv":
		return report.WriteCSV(f, rpt, false)
	}
	return nil
}

// parseSampleCSV parses the embedded sample CSV using the parser package logic.
func parseSampleCSV() ([]model.UsageItem, error) {
	// Write to a temp file and parse — avoids a dependency cycle
	// by reusing the same CSV field logic inline.
	tmp, err := os.CreateTemp("", "krr-demo-*.csv")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(SampleCSV); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()

	// Import parser lazily to avoid import cycle
	return parseSampleCSVFromFile(tmp.Name())
}
