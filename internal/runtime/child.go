package runtime

import (
	"fmt"
	"syscall"
)

func setupContainer(rootfs string) error {
	// Change hostname
	if err := syscall.Sethostname([]byte("container")); err != nil {
		return fmt.Errorf("set hostname: %w", err)
	}

	// Change root filesystem
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot: %w", err)
	}

	// Change to root directory
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	// Mount proc filesystem
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		if err != syscall.EBUSY {
			return fmt.Errorf("mount proc: %w", err)
		}
	}

	// Mount sys filesystem
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		if err != syscall.EBUSY {
			return fmt.Errorf("mount sys: %w", err)
		}
	}

	return nil
}