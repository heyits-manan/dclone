package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/heyits-manan/dclone/internal/cgroup"
)

func Run(rootfs, command string, args []string, memoryLimit string) error {
	// We're the parent - set up namespaces and re-exec
	return runParent(rootfs, command, args, memoryLimit)
}

func runParent(rootfs, command string, args []string, memoryLimit string) error {
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

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start child process: %w", err)
	}

	if memoryLimit != "" {
		cg, err := cgroup.NewMemoryCgroup(fmt.Sprintf("container-%d", cmd.Process.Pid), memoryLimit)
		if err != nil {
			return err
		}
		defer cg.Cleanup()

		if err := cg.AddProcess(cmd.Process.Pid); err != nil {
			return fmt.Errorf("add process to cgroup: %w", err)
		}
	}

	return cmd.Wait()
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