# krr-lite — Kubernetes Resource Request Recommender

krr-lite reads Kubernetes workload usage data from local CSV or JSON files and
recommends better CPU and memory resource requests.

It helps DevOps, SRE, FinOps, and platform engineers reduce over-provisioned
resources and detect under-provisioned workloads — without requiring live
cluster access, a Prometheus connection, or any cloud integration.

> **This tool is read-only. It does not apply changes.**
> Always review recommendations before updating manifests or Helm values.

---

## Why right-size resource requests?

Kubernetes schedules pods based on resource **requests**, not actual usage.
When requests are set too high:

- Nodes fill up faster than necessary, reducing cluster density.
- Cloud costs increase because node autoscaling triggers prematurely.
- Bin-packing efficiency drops across the cluster.

When requests are set too low or missing:

- Pods may be throttled (CPU) or OOM-killed (memory) under load.
- The scheduler places pods on nodes without enough headroom.
- Reliability decreases under peak load.

krr-lite takes pre-collected usage statistics (percentiles, max) and computes
safe, actionable recommendations per container.

---

## MVP scope

This MVP is local-file based, read-only, and deterministic.

**Supports:**
- CSV and JSON input files
- Three recommendation profiles (conservative, balanced, aggressive)
- Text, JSON, Markdown, and CSV output
- Validation of input data
- Risk and confidence ratings per recommendation
- Fail-on-risk exit code for CI pipelines

**Does not (yet) support:**
- Live Kubernetes API access
- Prometheus or Mimir integration
- Helm/Kustomize patch generation
- GitHub PR comment automation
- Cloud cost estimation
- Web UI or SaaS backend

---

## Install

### From source

```sh
git clone https://github.com/krr-lite/krr-lite
cd krr-lite
go build -o krr-lite .
```

### Run without installing

```sh
go run . <command>
```

---

## Quick start

```sh
# Create a demo directory with sample data and pre-generated reports
krr-lite init --output ./krr-demo

# Validate the sample input
krr-lite validate --file ./krr-demo/usage.csv

# Generate recommendations (text output, balanced profile)
krr-lite recommend --file ./krr-demo/usage.csv
```

---

## Commands

### `krr-lite version`

```sh
krr-lite version
```

### `krr-lite init`

Creates an example project directory with sample usage data and generated reports.

```sh
krr-lite init --output ./krr-demo
krr-lite init --output ./krr-demo --force   # overwrite if exists
```

Creates:

```
krr-demo/
  README.md
  usage.csv
  usage.json
  reports/
    recommendations.md
    recommendations.json
    recommendations.csv
```

### `krr-lite validate`

Validates a usage data file.

```sh
krr-lite validate --file ./usage.csv
krr-lite validate --file ./usage.json
krr-lite validate --file ./usage.csv --min-samples 50
```

Exits non-zero if there are validation errors. Warnings are printed but do not
cause a failure.

### `krr-lite recommend`

Reads usage data and generates CPU/memory request recommendations.

```sh
# Text output (default)
krr-lite recommend --file ./usage.csv

# Profile selection
krr-lite recommend --file ./usage.csv --profile conservative
krr-lite recommend --file ./usage.csv --profile aggressive

# Format and output file
krr-lite recommend --file ./usage.csv --format markdown --output recommendations.md
krr-lite recommend --file ./usage.csv --format json     --output recommendations.json
krr-lite recommend --file ./usage.csv --format csv      --output recommendations.csv

# Include no-change recommendations
krr-lite recommend --file ./usage.csv --include-unchanged

# CI: fail if any high-risk recommendation exists
krr-lite recommend --file ./usage.csv --fail-on-risk high

# Require more samples for high confidence
krr-lite recommend --file ./usage.csv --min-samples 100
```

---

## Input format

### Required columns / fields

| Field | Type | Description |
|---|---|---|
| namespace | string | Kubernetes namespace |
| workload_kind | string | e.g. Deployment, StatefulSet |
| workload_name | string | Workload name |
| container_name | string | Container name |
| current_cpu_request_mcores | float | Current CPU request in millicores |
| current_memory_request_mib | float | Current memory request in MiB |
| cpu_p50_mcores | float | CPU p50 usage in millicores |
| cpu_p95_mcores | float | CPU p95 usage in millicores |
| cpu_p99_mcores | float | CPU p99 usage in millicores |
| cpu_max_mcores | float | CPU max usage in millicores |
| memory_p50_mib | float | Memory p50 usage in MiB |
| memory_p95_mib | float | Memory p95 usage in MiB |
| memory_p99_mib | float | Memory p99 usage in MiB |
| memory_max_mib | float | Memory max usage in MiB |
| samples | int | Number of data samples |

### Optional columns / fields

| Field | Default |
|---|---|
| cluster | unknown |
| team | unknown |
| service | workload_name |
| environment | unknown |
| window | unknown |
| current_cpu_limit_mcores | 0 |
| current_memory_limit_mib | 0 |

### CSV example

```csv
cluster,namespace,workload_kind,workload_name,container_name,team,service,environment,current_cpu_request_mcores,current_memory_request_mib,current_cpu_limit_mcores,current_memory_limit_mib,cpu_p50_mcores,cpu_p95_mcores,cpu_p99_mcores,cpu_max_mcores,memory_p50_mib,memory_p95_mib,memory_p99_mib,memory_max_mib,samples,window
prod-kr,default,Deployment,api,api,platform,api,production,500,1024,1000,2048,80,160,220,300,300,520,650,800,720,7d
```

### JSON example

```json
{
  "items": [
    {
      "cluster": "prod-kr",
      "namespace": "default",
      "workload_kind": "Deployment",
      "workload_name": "api",
      "container_name": "api",
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
      "samples": 720
    }
  ]
}
```

---

## Recommendation profiles

| Profile | CPU formula | Memory formula |
|---|---|---|
| conservative | max(p95 × 1.5, p99 × 1.1) | max(p95 × 1.3, p99 × 1.1) |
| balanced _(default)_ | max(p95 × 1.25, p99) | max(p95 × 1.2, p99) |
| aggressive | p95 × 1.1 | p95 × 1.1 |

CPU recommendations round up to the nearest **10m**.
Memory recommendations round up to the nearest **16Mi**.
Minimum CPU: **10m**. Minimum memory: **32Mi**.

---

## Output formats

| Format | Flag | Notes |
|---|---|---|
| text | `--format text` | Default; human-readable terminal output |
| json | `--format json` | Full Report struct; always includes all recommendations |
| markdown | `--format markdown` | GitHub PR comment compatible |
| csv | `--format csv` | Spreadsheet-friendly |

---

## Risk levels

Each recommendation is assigned a risk level based on the findings:

| Risk | Conditions |
|---|---|
| high | Under-requested memory, recommendation below p99, or under-requested CPU below p95 |
| medium | Missing requests, low sample count, under-requested CPU, conflicting metrics |
| low | Over-requested CPU or memory |
| none | No issues found |

---

## Fail-on-risk (CI integration)

Use `--fail-on-risk` to exit non-zero when risky recommendations are found.
The report is always written before exiting.

| Value | Exits non-zero if |
|---|---|
| none | Never |
| low | Any low, medium, or high risk recommendation exists |
| medium | Any medium or high risk recommendation exists |
| high | Any high risk recommendation exists |

Example CI usage:

```sh
krr-lite recommend --file usage.csv --fail-on-risk high
```

---

## Limitations

- Does not connect to Kubernetes, Prometheus, or any external system.
- Does not apply or generate patches for manifests.
- Usage statistics must be pre-collected and exported by the user.
- Savings estimates are request-delta only; actual cloud cost impact varies.
- Does not account for VPA, KEDA, or HPA behaviour.

---

## Roadmap (not implemented)

- GitHub Actions integration
- GitHub PR comment bot
- Prometheus / Mimir live query
- Kubernetes API integration
- Helm values patch generation
- Kustomize patch generation
- VPA recommendation import/export
- Cloud pricing cost estimation
- Slack / Teams reports
- Web UI
