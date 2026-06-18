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
	bytes, err := parseMemoryLimit(limit)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(basePath, id)

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("create cgroup: %w", err)
	}

	memoryMaxPath := filepath.Join(path, "memory.max")
	if err := os.WriteFile(memoryMaxPath, []byte(strconv.FormatInt(bytes, 10)), 0644); err != nil {
		return nil, fmt.Errorf("set memory limit: %w", err)
	}

	return &MemoryCgroup{Path: path}, nil
}

func (cg *MemoryCgroup) AddProcess(pid int) error {
	procsPath := filepath.Join(cg.Path, "cgroup.procs")
	return os.WriteFile(procsPath, []byte(strconv.Itoa(pid)), 0644)
}

func (cg *MemoryCgroup) Cleanup() error {
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