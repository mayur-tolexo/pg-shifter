package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "PG-Shifter",
		Short: "This library make you enable migrate postgresql schema from golang struct",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
