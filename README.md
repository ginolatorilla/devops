# Gino's DevOps Tools

A collection of different tools that I use for DevOps-related tasks.

## Requirements

- [Go](https://go.dev/doc/install)
- Run `go run main.go check-requirements --list` for more details

## Usage

```shell
git clone https://github.com/ginolatorilla/devops.git
cd devops
make install
```

## Scripts

See `scripts` to learn more.

## Kubectl plugins

This tool installs the following `kubectl` plugins:

| Plugin                | Description                                                                                           |
| --------------------- | ----------------------------------------------------------------------------------------------------- |
| `list-addresses`      | Lists all IP addresses in the cluster                                                                 |
| `list-certs`          | Lists all certificates in the cluster and shows when they will be effective and when they will expire |
| `list-finalizers`     | Lists all Kubernetes resources that have finalizers                                                   |
| `list-unhealthy-pods` | Finds Kubernetes pods that are in a failed or unknown state                                           |
| `lookup-address`      | Finds Kubernetes resources by IP address                                                              |
| `trigger-cronjob`     | Launches a new job from an existing cronjob                                                           |
