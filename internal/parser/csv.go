package parser

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

// column name → field setter mappings for UsageItem
type csvField struct {
	required bool
	set      func(item *model.UsageItem, value string) error
}

var csvFields = map[string]csvField{
	"cluster":                    {false, func(i *model.UsageItem, v string) error { i.Cluster = v; return nil }},
	"namespace":                  {true, func(i *model.UsageItem, v string) error { i.Namespace = v; return nil }},
	"workload_kind":              {true, func(i *model.UsageItem, v string) error { i.WorkloadKind = v; return nil }},
	"workload_name":              {true, func(i *model.UsageItem, v string) error { i.WorkloadName = v; return nil }},
	"container_name":             {true, func(i *model.UsageItem, v string) error { i.ContainerName = v; return nil }},
	"team":                       {false, func(i *model.UsageItem, v string) error { i.Team = v; return nil }},
	"service":                    {false, func(i *model.UsageItem, v string) error { i.Service = v; return nil }},
	"environment":                {false, func(i *model.UsageItem, v string) error { i.Environment = v; return nil }},
	"window":                     {false, func(i *model.UsageItem, v string) error { i.Window = v; return nil }},
	"current_cpu_request_mcores": {true, floatSetter(func(i *model.UsageItem, v float64) { i.CurrentCPURequestMcores = v })},
	"current_memory_request_mib": {true, floatSetter(func(i *model.UsageItem, v float64) { i.CurrentMemoryRequestMiB = v })},
	"current_cpu_limit_mcores":   {false, floatSetter(func(i *model.UsageItem, v float64) { i.CurrentCPULimitMcores = v })},
	"current_memory_limit_mib":   {false, floatSetter(func(i *model.UsageItem, v float64) { i.CurrentMemoryLimitMiB = v })},
	"cpu_p50_mcores":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.CPUP50Mcores = v })},
	"cpu_p95_mcores":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.CPUP95Mcores = v })},
	"cpu_p99_mcores":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.CPUP99Mcores = v })},
	"cpu_max_mcores":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.CPUMaxMcores = v })},
	"memory_p50_mib":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.MemoryP50MiB = v })},
	"memory_p95_mib":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.MemoryP95MiB = v })},
	"memory_p99_mib":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.MemoryP99MiB = v })},
	"memory_max_mib":             {true, floatSetter(func(i *model.UsageItem, v float64) { i.MemoryMaxMiB = v })},
	"samples":                    {true, intSetter(func(i *model.UsageItem, v int) { i.Samples = v })},
}

func floatSetter(apply func(*model.UsageItem, float64)) func(*model.UsageItem, string) error {
	return func(item *model.UsageItem, value string) error {
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("expected number, got %q", value)
		}
		apply(item, v)
		return nil
	}
}

func intSetter(apply func(*model.UsageItem, int)) func(*model.UsageItem, string) error {
	return func(item *model.UsageItem, value string) error {
		v, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("expected integer, got %q", value)
		}
		apply(item, v)
		return nil
	}
}

// ParseCSV reads a CSV file and returns parsed usage items with defaults applied.
func ParseCSV(path string) ([]model.UsageItem, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(strings.NewReader(string(data)))
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read CSV header: %w", err)
	}

	// Normalise headers
	for i, h := range headers {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Check all required columns are present
	for col, f := range csvFields {
		if !f.required {
			continue
		}
		found := false
		for _, h := range headers {
			if h == col {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("CSV is missing required column %q", col)
		}
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("cannot read CSV records: %w", err)
	}

	var items []model.UsageItem
	for rowIdx, record := range records {
		if len(record) != len(headers) {
			return nil, fmt.Errorf("row %d: expected %d columns, got %d", rowIdx+2, len(headers), len(record))
		}
		var item model.UsageItem
		for colIdx, header := range headers {
			f, ok := csvFields[header]
			if !ok {
				// unknown column, skip
				continue
			}
			if err := f.set(&item, record[colIdx]); err != nil {
				return nil, fmt.Errorf("row %d, column %q: %w", rowIdx+2, header, err)
			}
		}
		item.ApplyDefaults()
		items = append(items, item)
	}
	return items, nil
}
