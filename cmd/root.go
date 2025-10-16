package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snip",
	Short: "Snip — get just the directory or file you need from a repository",
	Long:  "Snip downloads a single directory or file from a repository without cloning the entire repo.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
