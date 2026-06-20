package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/heyits-manan/dclone/cmd"
	"github.com/heyits-manan/dclone/internal/runtime"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// dclone runs the same binary in two modes:
	// 1. Normal user-facing CLI mode: `dclone run ...`
	// 2. Internal child mode: `dclone child <rootfs> <command> ...`
	//
	// The parent process starts the child by executing /proc/self/exe, which
	// points to this same binary. Passing "child" as argv[1] is how the new
	// process knows it should skip Cobra and perform container setup instead.
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// Internal argv shape:
		//   child <rootfs> <command> [args...]
		//
		// Example:
		//   child /tmp/alpine-rootfs /bin/sh
		rootfs := os.Args[2]
		command := os.Args[3]
		commandArgs := os.Args[4:]

		// Child mode is not a public CLI command. It is an implementation
		// detail used after the parent has created the new Linux namespaces.
		if err := runtime.RunChild(rootfs, command, commandArgs); err != nil {
			panic(err)
		}
		return
	}

	// Normal CLI path
	cmd.Execute()
}
