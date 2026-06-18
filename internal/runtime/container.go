package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/heyits-manan/dclone/internal/cgroup"
	"github.com/heyits-manan/dclone/internal/network"
)

func Run(rootfs, command string, args []string, memoryLimit string, networkConfig network.Config) error {
	return runParent(rootfs, command, args, memoryLimit, networkConfig)
}

func runParent(rootfs, command string, args []string, memoryLimit string, networkConfig network.Config) error {
	fmt.Printf("Running parent: %s %v\n", command, args)

	cmd := exec.Command("/proc/self/exe", append([]string{"child", rootfs, command}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
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

	if err := network.Setup(cmd.Process.Pid, networkConfig); err != nil {
		return fmt.Errorf("setup network: %w", err)
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
