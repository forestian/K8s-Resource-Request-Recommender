package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krr-lite/krr-lite/internal/model"
)

// Parse reads a usage file (CSV or JSON) and returns the items.
func Parse(path string) ([]model.UsageItem, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return ParseCSV(path)
	case ".json":
		return ParseJSON(path)
	default:
		return nil, fmt.Errorf("unsupported file extension %q: must be .csv or .json", ext)
	}
}

// readFile is a helper for reading a file with a clean error message.
func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %w", path, err)
	}
	return data, nil
}
