# krr-lite — Kubernetes Resource Request Recommender

Right-size your Kubernetes CPU and memory requests from pre-collected usage data — no cluster access required.

> **Read-only.** krr-lite never connects to Kubernetes or applies changes. Always review before updating manifests.

---

## Demo

GIF demo coming soon.

---

## Quick Demo

```sh
krr-lite recommend --file usage.csv
```

```
K8s Resource Request Recommender

Profile:        balanced
Input:          usage.csv
Items analyzed: 3

Summary:
  Set requests:      0
  Reduce requests:   2
  Increase requests: 0
  Review:            0
  No change:         1
  High risk:         0
  Medium risk:       0
  Low risk:          2
  Potential CPU request reduction:    960m
  Potential memory request reduction: 1312Mi

Recommendations:

[LOW] reduce_requests default/api container=api
CPU request:    500m -> 220m
Memory request: 1024Mi -> 656Mi
CPU status:     over_requested
Memory status:  over_requested
Confidence:     high

Reasons:
  - Current CPU request is more than 30% above recommended request.
  - Current memory request is more than 30% above recommended request.

Suggestions:
  - Review recent usage before applying.
  - Consider updating Helm values or manifests with the recommended requests.
  - Apply gradually and monitor throttling, OOM kills, and latency.

[LOW] reduce_requests default/worker container=worker
...
```

---

## Quick Start

### Download a prebuilt binary

Download from the [GitHub Releases page](https://github.com/forestian/K8s-Resource-Request-Recommender/releases).

**Linux / macOS:**
```sh
tar -xzf krr-lite_<version>_<os>_<arch>.tar.gz
chmod +x krr-lite
./krr-lite version
```

**Windows:**
```sh
# Extract the archive, then:
krr-lite.exe version
```

Prebuilt binaries: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`.
Each release includes `checksums.txt` for verification.

### Build from source

```sh
git clone https://github.com/forestian/K8s-Resource-Request-Recommender
cd K8s-Resource-Request-Recommender
go build -o krr-lite .
./krr-lite version
```

### Try the demo

```sh
krr-lite init --output ./krr-demo
krr-lite recommend --file ./krr-demo/usage.csv
```

---

## Use Cases

- Identify over-provisioned workloads to reduce cloud and node costs
- Detect under-provisioned containers before they cause OOM kills or throttling
- Enforce resource request policies in CI pipelines with `--fail-on-risk`
- Generate Markdown reports ready to paste into GitHub PR comments
- Audit resource requests across namespaces and teams from exported usage data

---

## Why right-size resource requests?

Kubernetes schedules pods based on resource **requests**, not actual usage.

When requests are set too high:
- Nodes fill up faster than necessary, reducing cluster density
- Cloud costs increase as node autoscaling triggers prematurely
- Bin-packing efficiency drops across the cluster

When requests are set too low or missing:
- Pods may be throttled (CPU) or OOM-killed (memory) under load
- The scheduler places pods on nodes without enough headroom
- Reliability decreases under peak load

krr-lite takes pre-collected usage statistics (percentiles + max) and computes safe, actionable recommendations per container.

---

## Commands

### `krr-lite version`

```sh
krr-lite version
```

### `krr-lite init`

Creates an example directory with sample data and pre-generated reports.

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

Exits non-zero on validation errors. Warnings are printed but do not cause failure.

### `krr-lite recommend`

Reads usage data and generates CPU/memory request recommendations.

```sh
# Text output (default, balanced profile)
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

# CI: fail with exit code 2 if any high-risk recommendation exists
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

## Example Output

### Text (default)

```
[LOW] reduce_requests default/api container=api
CPU request:    500m -> 220m
Memory request: 1024Mi -> 656Mi
CPU status:     over_requested
Memory status:  over_requested
Confidence:     high

Reasons:
  - Current CPU request is more than 30% above recommended request.
  - Current memory request is more than 30% above recommended request.

Suggestions:
  - Review recent usage before applying.
  - Consider updating Helm values or manifests with the recommended requests.
  - Apply gradually and monitor throttling, OOM kills, and latency.
```

### Markdown

```sh
krr-lite recommend --file usage.csv --format markdown --output recommendations.md
```

Generates a GitHub-flavored Markdown table suitable for PR comments or wikis.

### CSV

```sh
krr-lite recommend --file usage.csv --format csv --output recommendations.csv
```

One row per container; spreadsheet-friendly for bulk review.

---

## Risk levels

| Risk | Conditions |
|---|---|
| high | Under-requested memory, recommendation below p99, or CPU below p95 |
| medium | Missing requests, low sample count, under-requested CPU, conflicting metrics |
| low | Over-requested CPU or memory |
| none | No issues found |

---

## Fail-on-risk (CI integration)

Use `--fail-on-risk` to exit with code `2` when risky recommendations are found.
The report is always written before exiting.

| Value | Exits non-zero if |
|---|---|
| none | Never |
| low | Any low, medium, or high risk recommendation exists |
| medium | Any medium or high risk recommendation exists |
| high | Any high risk recommendation exists |

```sh
# In a CI pipeline — fail the job if any high-risk recommendation is found
krr-lite recommend --file usage.csv --fail-on-risk high
```

---

## Safety

- **Read-only:** krr-lite never connects to Kubernetes, Prometheus, or any external system.
- **No credentials:** does not read, store, or output secrets, tokens, or kubeconfig files.
- **Local only:** all processing happens on your machine with files you provide.
- Generated recommendations require human review before applying to production.

## Limitations

- Usage statistics must be pre-collected and exported by the user (no live scraping).
- Does not apply or generate patches for manifests, Helm values, or Kustomize overlays.
- Savings estimates are request-delta only; actual cloud cost impact varies by provider and pricing model.
- Does not account for VPA, KEDA, or HPA behavior.

---

## Roadmap

- GitHub Actions integration
- GitHub PR comment bot
- Prometheus / Mimir live query mode
- Helm values patch generation
- VPA recommendation import/export

---

## Part of the Forestian Cloud Native Toolkit

Part of the [Forestian Cloud Native Toolkit](https://github.com/forestian) — small CLI tools for Kubernetes, observability, GitOps, and platform engineering.
