# dclone

https://github.com/user-attachments/assets/36b417e8-4656-4916-bcaa-e376aba4c032

`dclone` is a minimal container runtime built in Go for learning how containers work under the hood.

This is not a Docker replacement. It is a small educational project focused on Linux processes, namespaces, `chroot`, mounts, cgroups, and basic container networking.

## What It Does

- Runs a command inside an Alpine root filesystem
- Creates a child process with Linux namespaces
- Isolates hostname, process IDs, and mount view
- Uses `chroot` so the child process sees the rootfs as `/`
- Mounts `/proc` and `/sys` inside the container
- Supports an experimental `--memory` flag using cgroups v2
- Connects the container to the host through a Linux bridge and veth pair

## Demo

```bash
sudo /tmp/dclone run /tmp/alpine-rootfs /bin/sh
```

Inside the container:

```bash
hostname
cat /etc/os-release
ls /
exit
```

Memory limit example:

```bash
sudo /tmp/dclone run --memory 64m /tmp/alpine-rootfs /bin/sh
```

Note: cgroups support depends on the Linux environment. In Lima on macOS, writing to `memory.max` may fail if the VM does not allow memory controller writes.

Network demo:

```bash
sudo /tmp/dclone run -p 8080:80 /tmp/alpine-rootfs /bin/sh
```

Inside the container:

```bash
mkdir -p /www
echo "hello from dclone container" > /www/index.html
httpd -f -p 80 -h /www
```

From another Lima terminal:

```bash
curl 10.88.0.2
```

Expected:

```text
hello from dclone container
```

Note: direct access to the container IP works. `localhost:8080` port publishing is still experimental.

## How It Works

```text
dclone run /tmp/alpine-rootfs /bin/sh
        |
        v
Parent process parses the CLI command
        |
        v
Parent starts a child process with new Linux namespaces
        |
        v
Child process re-enters the same binary using /proc/self/exe
        |
        v
Child sets hostname, chroots into the rootfs, mounts proc/sys
        |
        v
Parent connects the child network namespace to dclone0
        |
        v
/bin/sh runs inside the isolated environment
```

## Run Locally

This project must run on Linux. If you are on macOS, use a Linux VM such as Lima.

```bash
limactl start dclone-vm
limactl shell dclone-vm
```

Build:

```bash
cd /Users/itsmanan/Computer-Science/Backend/dclone
go build -o /tmp/dclone
```

Run:

```bash
sudo /tmp/dclone run /tmp/alpine-rootfs /bin/sh
```

## Tech Used

- Go
- Cobra
- Linux namespaces
- `chroot`
- Linux mounts
- cgroups v2
- Linux bridge and veth networking
- Lima for running Linux on macOS

## Current Limitations

- Educational project only
- No image pulling
- No daemon
- No container IDs or lifecycle tracking
- `localhost:8080` style port publishing is experimental
- `stop.go` exists as a placeholder and is not wired into the CLI

## Author

Manan
