# dclone

https://github.com/user-attachments/assets/36b417e8-4656-4916-bcaa-e376aba4c032

A minimal educational container runtime built in Go.

This project was built for learning how containers work under the hood. It is not intended to become a production runtime or a Docker replacement. The current scope is Phase 1: run a command inside a simple isolated Linux environment using namespaces, `chroot`, and mounts.

## Current Status

**Phase 1 Complete: Educational Container Execution**

✅ `dclone run <rootfs> <command> [args...]` — Run a container with isolated namespaces

### What Works Now
- **Process isolation** via PID namespace (container sees only its own processes)
- **Hostname isolation** via UTS namespace (container has its own hostname)
- **Filesystem isolation** via Mount namespace + chroot (container has its own root filesystem)
- **Re-exec pattern** — Parent process creates namespaces, child process enters container
- **Alpine Linux** containers can run successfully

### What Is Intentionally Left Out
- `stop` command lifecycle management
- cgroups and resource limits
- container networking
- image pulling/layer management
- daemon/client architecture
- production hardening and security controls

`cmd/stop.go` currently exists as a placeholder, but `stopCmd` is not registered in the root command, so `dclone stop` is not available yet.

## Quick Start

### Prerequisites
- Linux environment (VM, Lima, or bare metal)
- Go 1.21+
- Root privileges (for namespaces and chroot)

### Build

```bash
cd /Users/itsmanan/Computer-Science/Backend/dclone
go mod tidy
go build -o /tmp/dclone
```

### Download Alpine Rootfs

```bash
cd /tmp
wget https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/aarch64/alpine-minirootfs-3.19.0-aarch64.tar.gz
mkdir -p alpine-rootfs
cd alpine-rootfs
tar -xzf ../alpine-minirootfs-3.19.0-aarch64.tar.gz
```

### Run a Container

```bash
sudo /tmp/dclone run /tmp/alpine-rootfs /bin/sh
```

You should see a prompt like `/ #` indicating you're inside the Alpine container.

### Verify Container Isolation

Inside the container, run:
```bash
hostname          # Shows: container
ps aux            # Shows the isolated process namespace
ls /              # Shows Alpine filesystem (bin, etc, lib, usr, ...)
cat /etc/os-release  # Shows Alpine Linux
exit              # Exit the container
```

## Architecture

```
User types: dclone run /tmp/alpine-rootfs /bin/sh
     ↓
main.go
     ↓
cmd/run.go (Cobra CLI)
     ↓
runtime.Run() (parent process)
     ↓
Creates child process with namespaces:
  - CLONE_NEWUTS (hostname)
  - CLONE_NEWPID (process IDs)
  - CLONE_NEWNS (filesystem mounts)
     ↓
Re-executes: dclone child /tmp/alpine-rootfs /bin/sh
     ↓
runtime.RunChild() (child process)
     ↓
setupContainer():
  1. Sethostname("container")
  2. Chroot(/tmp/alpine-rootfs)
  3. Chdir("/")
  4. Mount /proc
  5. Mount /sys
     ↓
Run /bin/sh (user's command)
```

## Project Structure

```
dclone/
├── cmd/
│   ├── root.go          # Root command (dclone --help)
│   ├── run.go           # Run command: dclone run <rootfs> <command>
│   └── stop.go          # Placeholder; not registered yet
├── internal/
│   ├── runtime/
│   │   ├── container.go # Parent process: creates namespaces and re-execs
│   │   └── child.go     # Child process: chroot, mounts, and runs command
│   ├── namespace/
│   │   └── setup.go     # Placeholder for future namespace-specific operations
├── daemon/               # Placeholder for daemon
├── docs/                 # Placeholder for documentation
├── main.go              # Entry point, parent/child detection
├── go.mod               # Go dependencies
└── README.md            # This file
```

## Key Design Decisions

### Re-Exec Pattern
The same binary runs twice:
1. **Parent process**: Parses CLI, creates namespaces, forks child
2. **Child process**: Enters container, sets up environment, runs user's command

This is necessary because Linux namespaces can only be applied to **new processes**, not the current process.

### Educational Process Model
The parent stays outside the container-like environment and starts a child process with isolated namespaces. The child then sets up hostname, root filesystem, and mounts before running the requested command.

The current implementation uses `exec.Command(...).Run()` for the user's command. A more complete runtime would typically replace the child process with the user's command using `syscall.Exec` so the command has cleaner PID 1 semantics inside the container.

### Why chroot (not pivot_root)
- `chroot` is simpler to understand for learning purposes
- `pivot_root` is more secure but more complex
- We may migrate to `pivot_root` in later phases

### Error Handling
- Uses Go's idiomatic `error` return values
- Wraps errors with context using `fmt.Errorf("...: %w", err)`
- Uses `panic` for critical failures (educational code, not production)

## Milestones

### Milestone 1: Container Execution ✅
- `run` command
- PID namespace
- Mount namespace
- UTS namespace
- chroot filesystem isolation

### Future Learning Ideas
These are intentionally not implemented in the current learning version:

#### Resource Limits
- `stop` command
- Memory limits
- CPU limits
- cgroups v2

#### Image Management
- Image loading
- Layer management
- Rootfs preparation

#### Networking
- Bridge interface
- NAT
- Port forwarding
- veth pairs
- Network namespaces

#### Daemon Mode
- Client-server architecture
- REST API
- Container lifecycle management

## Development

### Lima Workflow (macOS)
Since container runtimes require Linux kernel features, we develop on macOS but build and test in a Lima VM:

```bash
# On Mac: edit files in your IDE
# In Lima VM:
cd /Users/itsmanan/Computer-Science/Backend/dclone
go build -o /tmp/dclone
sudo /tmp/dclone run /tmp/alpine-rootfs /bin/sh
```

### Git Commits
We follow small, focused commits:
```
feat: add CLI commands with cobra
feat: add runtime package for container execution
fix: add missing imports to main.go
fix: handle EBUSY errors for proc/sys mounts
```

## Tech Stack

- **Go** — Programming language
- **Cobra** — CLI framework
- **Linux syscalls** — namespaces, chroot, mounts
- **cgroup v2** — Resource limits (planned)
- **Lima** — Linux VM for macOS development

## Learning Goals

This project is built to understand:
- How Docker actually works under the hood
- Linux namespaces and process isolation
- cgroups for resource management
- Container networking
- The difference between containers and VMs

## Non-Goals

- Production use
- Orchestration (Kubernetes replacement)
- Distributed scheduling
- GUI
- Windows/macOS container support

## License

MIT

## Author

Manan

---

*Built for learning. Not for production.*
