package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/krr-lite/krr-lite/internal/model"
)

// WriteJSON writes the full report as JSON to w.
// JSON always includes all recommendations regardless of includeUnchanged.
func WriteJSON(w io.Writer, report *model.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("cannot encode report as JSON: %w", err)
	}
	return nil
}
