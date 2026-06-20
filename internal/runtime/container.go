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

	// Re-exec the current binary as the container child process.
	//
	// /proc/self/exe is a Linux procfs symlink to the executable of the
	// currently running process. Using it avoids needing to know where the
	// dclone binary lives on disk. The child receives a private internal
	// command shape:
	//   child <rootfs> <command> [args...]
	//
	// The new process is still dclone, but main.go detects argv[1] == "child"
	// and jumps into RunChild instead of parsing the public Cobra CLI.
	cmd := exec.Command("/proc/self/exe", append([]string{"child", rootfs, command}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// These clone flags ask Linux to create new namespaces for the child.
	// The parent remains in the host/Lima namespaces and can manage the child.
	//
	// CLONE_NEWUTS: separate hostname/domainname view.
	// CLONE_NEWPID: separate process ID view for descendants.
	// CLONE_NEWNS:  separate mount table, so container mounts do not affect host.
	// CLONE_NEWNET: separate network stack, later connected with bridge/veth.
	//
	// PID namespaces are a key reason for parent/child separation: the current
	// process cannot simply become PID 1 in a new PID namespace. Linux applies
	// the new PID namespace to a newly created child process.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
	}

	// Start gives us the child PID before waiting. That PID is needed for
	// cgroups and networking setup:
	// - cgroups: write the PID into cgroup.procs
	// - networking: move a veth peer into the child's network namespace
	//
	// cmd.Run() would start and wait in one call, which is too late for this
	// setup work.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start child process: %w", err)
	}

	if memoryLimit != "" {
		// cgroups v2 resource limits are controlled through files under
		// /sys/fs/cgroup. This creates a cgroup for the process, writes the
		// memory limit to memory.max, then adds the child PID to cgroup.procs.
		cg, err := cgroup.NewMemoryCgroup(fmt.Sprintf("container-%d", cmd.Process.Pid), memoryLimit)
		if err != nil {
			return err
		}
		defer cg.Cleanup()

		if err := cg.AddProcess(cmd.Process.Pid); err != nil {
			return fmt.Errorf("add process to cgroup: %w", err)
		}
	}

	// If networking is enabled, connect the child's isolated network namespace
	// back to the host through a Linux bridge and veth pair. This is what makes
	// `curl 10.88.0.2` from the host reach a server running inside the child.
	if err := network.Setup(cmd.Process.Pid, networkConfig); err != nil {
		return fmt.Errorf("setup network: %w", err)
	}

	// Wait keeps the parent alive until the container command exits. In this
	// simple runtime, container lifetime equals the lifetime of the foreground
	// command such as /bin/sh or httpd.
	return cmd.Wait()
}

func RunChild(rootfs, command string, args []string) error {
	fmt.Printf("Running child: %s %v\n", command, args)

	// This runs inside the child process after the parent has created the new
	// namespaces. From here onward, changes like hostname, chroot, and mounts
	// affect the child/container environment instead of the parent process.
	if err := setupContainer(rootfs); err != nil {
		return fmt.Errorf("setup container: %w", err)
	}

	// Run the user's command inside the prepared environment. A production
	// runtime would often use syscall.Exec to replace this child process with
	// the user's command; this learning runtime keeps exec.Command for clarity.
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
