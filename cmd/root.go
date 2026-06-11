package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dclone",
	Short: "A minimal container runtime",
	Long:  `dclone is a simple container runtime for learning Linux namespaces, cgroups, and networking.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}