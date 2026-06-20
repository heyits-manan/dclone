package runtime

import (
	"fmt"
	"syscall"
)

func setupContainer(rootfs string) error {
	// The UTS namespace gives this child its own hostname view. Setting the
	// hostname here should not rename the Lima VM/host; it only changes what
	// processes inside this namespace see when they run `hostname`.
	if err := syscall.Sethostname([]byte("container")); err != nil {
		return fmt.Errorf("set hostname: %w", err)
	}

	// chroot changes what "/" means for this process and its children.
	// If rootfs is /tmp/alpine-rootfs, then /bin/sh inside the container
	// resolves to /tmp/alpine-rootfs/bin/sh from the host's point of view.
	//
	// chroot does not create a VM and does not provide a kernel. It only changes
	// filesystem path resolution for this process tree.
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot: %w", err)
	}

	// After chroot, move the current working directory to the new "/".
	// Without this, the process may still have an old working directory handle
	// from before the chroot.
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	// /proc is a virtual filesystem provided by the Linux kernel. Tools like
	// ps read it to understand running processes. Mounting it after chroot
	// gives the container a /proc path inside its new root filesystem.
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		// EBUSY means something is already mounted at this target. For this
		// learning runtime, treating an existing mount as acceptable makes
		// repeated experiments less fragile.
		if err != syscall.EBUSY {
			return fmt.Errorf("mount proc: %w", err)
		}
	}

	// /sys is another kernel-backed virtual filesystem. It exposes system and
	// device information. Like /proc, it needs to be mounted inside the chroot
	// if commands inside the container expect it to exist.
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		if err != syscall.EBUSY {
			return fmt.Errorf("mount sys: %w", err)
		}
	}

	return nil
}
