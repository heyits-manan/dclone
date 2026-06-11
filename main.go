package main

import (
	"math/rand"
	"time"

	"github.com/heyits-manan/dclone/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Check if we're the child process (re-execed by parent)
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// Extract args: child <rootfs> <command> [args...]
		rootfs := os.Args[2]
		command := os.Args[3]
		commandArgs := os.Args[4:]

		// Run child logic directly, bypass Cobra
		if err := runtime.RunChild(rootfs, command, commandArgs); err != nil {
			panic(err)
		}
		return
	}

	// Normal CLI path
	cmd.Execute()
}