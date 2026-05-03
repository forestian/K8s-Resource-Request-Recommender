package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current tool version.
const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("krr-lite version %s\n", Version)
	},
}
