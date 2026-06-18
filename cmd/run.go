package cmd

import (
	"fmt"
	"os"

	"github.com/heyits-manan/dclone/internal/network"
	"github.com/heyits-manan/dclone/internal/runtime"
	"github.com/spf13/cobra"
)

var memoryLimit string
var publishPort string

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

		networkConfig, err := network.ParsePublish(publishPort)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runtime.Run(rootfs, command, commandArgs, memoryLimit, networkConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().StringVar(&memoryLimit, "memory", "", "memory limit, for example: 64m, 128m, 1g")
	runCmd.Flags().StringVarP(&publishPort, "publish", "p", "", "publish a port, for example: 8080:80")
}
