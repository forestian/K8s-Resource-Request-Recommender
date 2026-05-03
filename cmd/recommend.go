package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krr-lite/krr-lite/internal/parser"
	"github.com/krr-lite/krr-lite/internal/recommend"
	"github.com/krr-lite/krr-lite/internal/report"
	"github.com/krr-lite/krr-lite/internal/validate"
	"github.com/spf13/cobra"
)

var (
	recFile             string
	recOutput           string
	recFormat           string
	recProfile          string
	recMinSamples       int
	recIncludeUnchanged bool
	recFailOnRisk       string
	recForce            bool
)

var validFormats = map[string]bool{
	"text": true, "json": true, "markdown": true, "csv": true,
}

var validFailOnRisk = map[string]bool{
	"none": true, "low": true, "medium": true, "high": true,
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Generate CPU and memory request recommendations from usage data",
	Long: `Reads a CSV or JSON usage file, validates it, and outputs resource
request recommendations with risk and confidence ratings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRecommend()
	},
}

func init() {
	recommendCmd.Flags().StringVar(&recFile, "file", "", "Input usage file (.csv or .json)")
	recommendCmd.Flags().StringVar(&recOutput, "output", "", "Output file path (default: stdout)")
	recommendCmd.Flags().StringVar(&recFormat, "format", "text", "Output format: text, json, markdown, csv")
	recommendCmd.Flags().StringVar(&recProfile, "profile", "balanced", "Recommendation profile: conservative, balanced, aggressive")
	recommendCmd.Flags().IntVar(&recMinSamples, "min-samples", 30, "Minimum samples required for high confidence")
	recommendCmd.Flags().BoolVar(&recIncludeUnchanged, "include-unchanged", false, "Include no_change recommendations in output")
	recommendCmd.Flags().StringVar(&recFailOnRisk, "fail-on-risk", "none", "Exit non-zero if risk meets threshold: none, low, medium, high")
	recommendCmd.Flags().BoolVar(&recForce, "force", false, "Overwrite existing output file")
}

func runRecommend() error {
	// Validate flags
	if recFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --file is required")
		os.Exit(1)
	}

	if _, err := os.Stat(recFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file %q does not exist\n", recFile)
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(recFile))
	if ext != ".csv" && ext != ".json" {
		fmt.Fprintf(os.Stderr, "Error: file must be .csv or .json, got %q\n", ext)
		os.Exit(1)
	}

	if !validFormats[recFormat] {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q; must be one of: text, json, markdown, csv\n", recFormat)
		os.Exit(1)
	}

	if !recommend.IsValidProfile(recProfile) {
		fmt.Fprintf(os.Stderr, "Error: invalid profile %q; must be one of: conservative, balanced, aggressive\n", recProfile)
		os.Exit(1)
	}

	if recMinSamples < 1 {
		fmt.Fprintln(os.Stderr, "Error: --min-samples must be >= 1")
		os.Exit(1)
	}

	if !validFailOnRisk[recFailOnRisk] {
		fmt.Fprintf(os.Stderr, "Error: invalid --fail-on-risk %q; must be one of: none, low, medium, high\n", recFailOnRisk)
		os.Exit(1)
	}

	// Overwrite protection
	if recOutput != "" {
		if _, err := os.Stat(recOutput); err == nil && !recForce {
			fmt.Fprintf(os.Stderr, "Error: output file %q already exists; use --force to overwrite\n", recOutput)
			os.Exit(1)
		}
	}

	// Parse
	items, err := parser.Parse(recFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate (print warnings, fail on errors)
	result := validate.Validate(items, recMinSamples)
	if !result.OK() {
		fmt.Fprintf(os.Stderr, "Validation failed:\n")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  ERROR: %s\n", e)
		}
		os.Exit(1)
	}

	// Generate recommendations
	opts := recommend.Options{
		Profile:    recProfile,
		MinSamples: recMinSamples,
	}
	rpt, _ := recommend.Recommend(items, recFile, opts)

	// Determine output writer
	out := os.Stdout
	if recOutput != "" {
		f, err := os.Create(recOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot create output file %q: %v\n", recOutput, err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	// Write report
	switch recFormat {
	case "text":
		if err := report.WriteText(out, rpt, recIncludeUnchanged); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(1)
		}
	case "json":
		if err := report.WriteJSON(out, rpt); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(1)
		}
	case "markdown":
		if err := report.WriteMarkdown(out, rpt, recIncludeUnchanged); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(1)
		}
	case "csv":
		if err := report.WriteCSV(out, rpt, recIncludeUnchanged); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(1)
		}
	}

	// Fail-on-risk: check after writing the report
	if recFailOnRisk != "none" && recFailOnRisk != "" {
		for _, rec := range rpt.Recommendations {
			if recommend.RiskMeetsThreshold(rec.Risk, recFailOnRisk) {
				fmt.Fprintf(os.Stderr, "\nFailed: recommendation for %s/%s/%s has risk %q which meets --fail-on-risk threshold %q\n",
					rec.Namespace, rec.WorkloadName, rec.ContainerName, rec.Risk, recFailOnRisk)
				os.Exit(2)
			}
		}
	}

	return nil
}
