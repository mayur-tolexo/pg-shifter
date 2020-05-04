package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "ct",
	Short: "Create Table",
	Long: `This will allow you to create table. Pass table names space separeted.
For creating all tables pass all as argument
i.e ./shifter ct all`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD", args)
	},
}
