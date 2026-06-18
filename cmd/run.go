package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/heyits-manan/dclone/internal/runtime"
)

var memoryLimit string

var runCmd = &cobra.Command{
	Use:   "run [rootfs] [command] [args...]",
	Short: "Run a container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		rootfs := args[0]
		command := args[1]
		commandArgs := args[2:]

		if _, err := os.Stat(rootfs); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: rootfs %s does not exist\n", rootfs)
			os.Exit(1)
		}

		if err := runtime.Run(rootfs, command, commandArgs, memoryLimit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init(){
	runCmd.Flags().StringVar(&memoryLimit, "memory", "", "memory limit, for example: 64m, 128m, 1g")
}