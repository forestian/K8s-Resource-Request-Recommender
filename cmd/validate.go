package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krr-lite/krr-lite/internal/parser"
	"github.com/krr-lite/krr-lite/internal/validate"
	"github.com/spf13/cobra"
)

var validateFile string
var validateMinSamples int

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a usage data file",
	Long:  `Checks a CSV or JSON usage file for structural and logical correctness.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if validateFile == "" {
			return fmt.Errorf("--file is required")
		}

		// File must exist
		if _, err := os.Stat(validateFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file %q does not exist\n", validateFile)
			os.Exit(1)
		}

		// Extension must be .csv or .json
		ext := strings.ToLower(filepath.Ext(validateFile))
		if ext != ".csv" && ext != ".json" {
			fmt.Fprintf(os.Stderr, "Error: file must be .csv or .json, got %q\n", ext)
			os.Exit(1)
		}

		items, err := parser.Parse(validateFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result := validate.Validate(items, validateMinSamples)

		if len(result.Warnings) > 0 {
			fmt.Printf("Warnings (%d):\n", len(result.Warnings))
			for _, w := range result.Warnings {
				fmt.Printf("  WARNING: %s\n", w)
			}
			fmt.Println()
		}

		if !result.OK() {
			fmt.Printf("Validation failed (%d error(s)):\n", len(result.Errors))
			for _, e := range result.Errors {
				fmt.Printf("  ERROR: %s\n", e)
			}
			os.Exit(1)
		}

		fmt.Printf("OK: %d items validated (%d warning(s))\n", len(items), len(result.Warnings))
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateFile, "file", "", "Input usage file (.csv or .json)")
	validateCmd.Flags().IntVar(&validateMinSamples, "min-samples", 30, "Minimum samples for high confidence")
}
