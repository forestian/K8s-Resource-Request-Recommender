package cmd

import (
	"fmt"
	"os"

	"github.com/krr-lite/krr-lite/internal/initdemo"
	"github.com/spf13/cobra"
)

var initOutputDir string
var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an example project directory with sample data and reports",
	Long: `Creates a demo directory with sample usage.csv, usage.json, and
pre-generated recommendation reports. Use this to explore the tool format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if initOutputDir == "" {
			return fmt.Errorf("--output is required")
		}
		if err := initdemo.Generate(initOutputDir, initForce); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initOutputDir, "output", "", "Output directory to create (required)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing files")
}
