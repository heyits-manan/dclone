package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const basePath = "/sys/fs/cgroup/dclone"

type MemoryCgroup struct {
	Path string
}

func NewMemoryCgroup(id string, limit string) (*MemoryCgroup, error) {
	// cgroups v2 accepts memory.max as a number of bytes. The CLI accepts
	// human-friendly values like 64m or 1g, so convert that first.
	bytes, err := parseMemoryLimit(limit)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(basePath, id)

	// Each container gets its own cgroup directory. In cgroups v2, creating a
	// directory under /sys/fs/cgroup creates a new resource-control group.
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("create cgroup: %w", err)
	}

	// memory.max is the cgroups v2 file that controls the maximum memory this
	// group can use. Writing "67108864" means 64 MiB.
	memoryMaxPath := filepath.Join(path, "memory.max")
	if err := os.WriteFile(memoryMaxPath, []byte(strconv.FormatInt(bytes, 10)), 0644); err != nil {
		return nil, fmt.Errorf("set memory limit: %w", err)
	}

	return &MemoryCgroup{Path: path}, nil
}

func (cg *MemoryCgroup) AddProcess(pid int) error {
	// cgroup.procs is how a process joins a cgroup. Writing the child PID here
	// tells Linux that this process should be controlled by this cgroup's
	// limits, including memory.max.
	procsPath := filepath.Join(cg.Path, "cgroup.procs")
	return os.WriteFile(procsPath, []byte(strconv.Itoa(pid)), 0644)
}

func (cg *MemoryCgroup) Cleanup() error {
	// Best-effort cleanup after the foreground container exits. This will only
	// succeed once Linux has removed all processes from the cgroup.
	return os.Remove(cg.Path)
}

func parseMemoryLimit(value string) (int64, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return 0, fmt.Errorf("memory limit cannot be empty")
	}

	multiplier := int64(1)

	switch {
	case strings.HasSuffix(value, "k"):
		multiplier = 1024
		value = strings.TrimSuffix(value, "k")
	case strings.HasSuffix(value, "m"):
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(value, "m")
	case strings.HasSuffix(value, "g"):
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(value, "g")
	}

	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory limit: %w", err)
	}

	if number <= 0 {
		return 0, fmt.Errorf("memory limit must be greater than zero")
	}

	return number * multiplier, nil
}
