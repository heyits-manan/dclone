package network

import (
	"fmt"
	"os/exec"
	"strings"
)

type Config struct {
	Enabled       bool
	HostPort      string
	ContainerPort string
	BridgeName    string
	ContainerIP   string
}

func ParsePublish(value string) (Config, error) {
	if value == "" {
		return Config{Enabled: false}, nil
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return Config{}, fmt.Errorf("publish must be in hostPort:containerPort format")
	}

	return Config{
		Enabled:       true,
		HostPort:      parts[0],
		ContainerPort: parts[1],
		BridgeName:    "dclone0",
		ContainerIP:   "10.88.0.2",
	}, nil
}

func Setup(pid int, cfg Config) error {
	if !cfg.Enabled {
		return nil
	}

	hostVeth := fmt.Sprintf("veth%d", pid)
	containerVeth := fmt.Sprintf("eth%d", pid)

	commands := [][]string{
		{"ip", "link", "add", cfg.BridgeName, "type", "bridge"},
		{"ip", "addr", "add", "10.88.0.1/24", "dev", cfg.BridgeName},
		{"ip", "link", "set", cfg.BridgeName, "up"},

		{"ip", "link", "add", hostVeth, "type", "veth", "peer", "name", containerVeth},
		{"ip", "link", "set", hostVeth, "master", cfg.BridgeName},
		{"ip", "link", "set", hostVeth, "up"},
		{"ip", "link", "set", containerVeth, "netns", fmt.Sprintf("%d", pid)},

		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "addr", "add", cfg.ContainerIP + "/24", "dev", containerVeth},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", containerVeth, "name", "eth0"},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "eth0", "up"},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "lo", "up"},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "route", "add", "default", "via", "10.88.0.1"},

		{"iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "--dport", cfg.HostPort, "-j", "DNAT", "--to-destination", cfg.ContainerIP + ":" + cfg.ContainerPort},
		{"iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", cfg.HostPort, "-j", "DNAT", "--to-destination", cfg.ContainerIP + ":" + cfg.ContainerPort},
		{"iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.88.0.0/24", "-j", "MASQUERADE"},
	}

	for _, command := range commands {
		if err := run(command[0], command[1:]...); err != nil {
			return err
		}
	}

	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}
