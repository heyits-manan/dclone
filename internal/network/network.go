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

	// This project uses one simple fixed container network:
	//
	//   host/Lima namespace:
	//     dclone0 bridge -> 10.88.0.1/24
	//
	//   container namespace:
	//     eth0           -> 10.88.0.2/24
	//
	// The -p flag is parsed for Docker-like CLI shape, but localhost port
	// publishing is still experimental. The currently verified path is direct
	// host-to-container access with `curl 10.88.0.2`.
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

	// A veth pair is a virtual Ethernet cable with two ends. Packets entering
	// one end come out the other.
	//
	// hostVeth stays in the host/Lima network namespace and gets attached to
	// the dclone0 bridge. containerVeth is moved into the child process's
	// network namespace, then renamed to eth0 so it looks like a normal
	// container network interface.
	hostVeth := fmt.Sprintf("veth%d", pid)
	containerVeth := fmt.Sprintf("eth%d", pid)

	commands := [][]string{
		// Create the host-side bridge. A Linux bridge behaves like a virtual
		// switch. dclone0 is the host-side connection point for containers.
		{"ip", "link", "add", cfg.BridgeName, "type", "bridge"},

		// Give the bridge an IP address on the private dclone subnet. From the
		// container's point of view, 10.88.0.1 acts like the gateway.
		{"ip", "addr", "add", "10.88.0.1/24", "dev", cfg.BridgeName},
		{"ip", "link", "set", cfg.BridgeName, "up"},

		// Create the virtual cable, attach the host end to the bridge, and turn
		// the host end on.
		{"ip", "link", "add", hostVeth, "type", "veth", "peer", "name", containerVeth},
		{"ip", "link", "set", hostVeth, "master", cfg.BridgeName},
		{"ip", "link", "set", hostVeth, "up"},

		// Move the container end of the veth pair into the child's network
		// namespace. After this command, the host namespace no longer sees that
		// interface directly.
		{"ip", "link", "set", containerVeth, "netns", fmt.Sprintf("%d", pid)},

		// nsenter runs the following `ip ...` commands inside the child's
		// network namespace. This is how the parent configures the interface
		// after moving it away from the host namespace.
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "addr", "add", cfg.ContainerIP + "/24", "dev", containerVeth},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", containerVeth, "name", "eth0"},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "eth0", "up"},
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "lo", "up"},

		// Add a default route inside the container. If the container sends
		// traffic outside 10.88.0.0/24, it should go through the bridge IP.
		{"nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "route", "add", "default", "via", "10.88.0.1"},

		// NAT rules for Docker-like port publishing. These are intentionally
		// experimental in this learning project: direct container IP access
		// works (`curl 10.88.0.2`), while localhost publishing depends on the
		// host/VM iptables behavior.
		{"iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "--dport", cfg.HostPort, "-j", "DNAT", "--to-destination", cfg.ContainerIP + ":" + cfg.ContainerPort},
		{"iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", cfg.HostPort, "-j", "DNAT", "--to-destination", cfg.ContainerIP + ":" + cfg.ContainerPort},
		{"iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.88.0.0/24", "-j", "MASQUERADE"},
	}

	for _, command := range commands {
		// Each entry stores the executable name at index 0 and its arguments
		// after that. exec.Command wants those as separate parameters:
		//   exec.Command("ip", "link", "set", ...)
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
		// Include command output in errors because Linux networking tools often
		// explain the real problem there, for example "File exists" or
		// "Cannot open network namespace".
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}
