package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func Run(rootfs, command string, args []string) error {
	// We're the parent - set up namespaces and re-exec
	return runParent(rootfs, command, args)
}

func runParent(rootfs, command string, args []string) error {
	fmt.Printf("Running parent: %s %v\n", command, args)

	cmd := exec.Command("/proc/self/exe", append([]string{"child", rootfs, command}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | 
                    syscall.CLONE_NEWPID | 
                    syscall.CLONE_NEWNS,
	}

	return cmd.Run()
}

func RunChild(rootfs, command string, args []string) error {
	fmt.Printf("Running child: %s %v\n", command, args)

	// Set up container environment
	if err := setupContainer(rootfs); err != nil {
		return fmt.Errorf("setup container: %w", err)
	}

	// Run the actual command
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}