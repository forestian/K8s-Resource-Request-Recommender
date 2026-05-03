package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "krr-lite",
	Short: "Kubernetes Resource Request Recommender",
	Long: `krr-lite analyzes Kubernetes workload usage data from local CSV or JSON
files and recommends better CPU and memory resource requests.

It helps DevOps, SRE, and platform engineers reduce over-requested resources
and detect under-requested workloads without requiring live cluster access.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(recommendCmd)
}
