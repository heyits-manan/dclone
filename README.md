# MyDocker Design Doc

## Goal

Build a minimal Docker clone to learn:

- namespaces
- cgroups
- networking
- image layers

## Non-goals

- orchestration
- distributed scheduling
- GUI

## Milestone 1

Container execution

Features:

- run
- stop

Linux concepts:

- PID namespace
- mount namespace

## Milestone 2

Filesystem isolation

Features:

- rootfs
- image loading

Linux concepts:

- chroot
- pivot_root

## Milestone 3

Resource limits

Features:

- memory limits
- cpu limits

Linux concepts:

- cgroups

## Milestone 4

Networking

Features:

- bridge
- NAT
- port forwarding

Linux concepts:

- veth pairs
- network namespaces

## Architecture

CLI
|
REST API
|
Daemon
|

- Runtime
- Storage
- Networking
- Cgroups
