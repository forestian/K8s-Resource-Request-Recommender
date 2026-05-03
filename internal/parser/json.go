package parser

import (
	"encoding/json"
	"fmt"

	"github.com/krr-lite/krr-lite/internal/model"
)

// ParseJSON reads a JSON file and returns parsed usage items with defaults applied.
func ParseJSON(path string) ([]model.UsageItem, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var file model.UsageFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("cannot parse JSON file %q: %w", path, err)
	}

	for i := range file.Items {
		file.Items[i].ApplyDefaults()
	}
	return file.Items, nil
}
