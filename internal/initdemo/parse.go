package initdemo

import (
	"github.com/krr-lite/krr-lite/internal/model"
	"github.com/krr-lite/krr-lite/internal/parser"
)

func parseSampleCSVFromFile(path string) ([]model.UsageItem, error) {
	return parser.ParseCSV(path)
}
