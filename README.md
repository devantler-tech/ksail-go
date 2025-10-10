[![codecov](https://codecov.io/gh/devantler-tech/ksail-go/graph/badge.svg?token=HSUfhaiXwq)](https://codecov.io/gh/devantler-tech/ksail-go)
[![CI - Go](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml)

# ksail

> [!IMPORTANT]
> This is a work in progress to migrate KSail to a Golang. This is a huge endeavour, but being able to leverage the power of the Go ecosystem will be invaluable. The amount of packages available in Go to support this project is immense, so switching programming language has the potential to greatly enhance the functionality, performance and ease of use of KSail. I also hope switching will promote adoption and contributions.

## Why KSail?

Setting up local Kubernetes environments shouldn't require juggling multiple CLIs, memorizing complex commands, or maintaining fragmented scripts. Whether you're developing applications, testing infrastructure, or learning Kubernetes, you need a fast, consistent way to spin up clusters and manage workloads—without the chaos.

**The Problem:**

- Installing and configuring Kind, K3d, or EKS clusters requires different tools and workflows
- Managing cluster lifecycle (create, start, stop, delete) means learning multiple CLI syntaxes
- Deploying workloads involves manual kubectl commands and tracking numerous YAML files
- Switching between clusters and contexts is error-prone
- Local development lacks the declarative simplicity Kubernetes promises

**The Solution:**

KSail provides a unified, declarative interface that wraps the Kubernetes tools you already trust. One command-line tool, consistent syntax, multiple distributions—from local development to cloud deployment.

## What is KSail?

KSail is a **Kubernetes cluster management CLI** that simplifies local development and testing. It's your all-in-one tool for:

✨ **Key Features:**

- 🚀 **Multi-Distribution Support** - Seamlessly work with Kind, K3d, and EKS clusters from a single interface
- 📦 **Declarative Configuration** - Define your entire cluster and workload setup in simple YAML files
- ⚡ **Fast Cluster Lifecycle** - Create, start, stop, and destroy clusters with single commands
- 🔄 **Workload Management** - Deploy and manage Kubernetes workloads with built-in reconciliation
- 🔐 **Secrets Management** - Integrated SOPS cipher support for encrypted configuration files
- 🎯 **Project Scaffolding** - Initialize new projects with best-practice structures
- 🔍 **Interactive Monitoring** - Built-in k9s integration for cluster visualization
- 🛠️ **Developer-First** - Designed for local development, testing, and CI/CD workflows

**Use Cases:**

- Local Kubernetes development and testing
- Multi-cluster development environments
- CI/CD pipeline testing
- Infrastructure-as-Code validation
- Kubernetes learning and experimentation
- Microservices development

## Getting Started

### Prerequisites

**System Requirements:**

- **Operating System**: Linux (amd64/arm64) or MacOS (amd64/arm64)
- **Docker**: Required for Kind and K3d clusters
- **Go 1.23.9+**: Only needed if building from source

**Optional Tools:**

- **kubectl**: For manual cluster interaction (KSail handles most operations)
- **k9s**: For terminal-based cluster UI (KSail includes built-in integration)

### Installation

#### Homebrew

Coming soon! Watch this repository for updates.

#### Go Install

Install the latest version directly with Go:

```bash
go install github.com/devantler-tech/ksail-go@latest
ksail --help
```

#### From Source

Build from source for the latest development version:

```bash
git clone https://github.com/devantler-tech/ksail-go.git
cd ksail-go
go build -o ksail .
./ksail --help
```

### Quick Start

Get a Kubernetes cluster running in seconds:

```bash
# Initialize a new project with Kind
ksail init --distribution Kind

# Create and start the cluster
ksail cluster create

# Check cluster status
ksail cluster info

# Connect to cluster with k9s
ksail cluster connect

# When done, tear down the cluster
ksail cluster delete
```

## Usage

### Basic Workflow

KSail follows a simple, intuitive workflow:

```bash
# 1. Initialize a project
ksail init --distribution Kind --output ./my-project
cd my-project

# 2. Create a cluster
ksail cluster create

# 3. Check cluster status
ksail cluster info
ksail cluster list

# 4. Deploy workloads
ksail workload reconcile

# 5. Interact with cluster
ksail cluster connect  # Opens k9s

# 6. Clean up
ksail cluster delete
```

### Cluster Management

**Supported Distributions:**

```bash
# Kind (Kubernetes in Docker) - Default
ksail init --distribution Kind

# K3d (K3s in Docker) - Lightweight
ksail init --distribution K3d

# EKS (Amazon Elastic Kubernetes Service)
ksail init --distribution EKS
```

**Lifecycle Commands:**

```bash
# Create a new cluster
ksail cluster create

# Start a stopped cluster
ksail cluster start

# Stop a running cluster
ksail cluster stop

# Delete a cluster
ksail cluster delete

# List all clusters
ksail cluster list

# View cluster information
ksail cluster info

# Connect with k9s
ksail cluster connect
```

### Workload Management

```bash
# Reconcile all workloads from k8s directory
ksail workload reconcile

# Apply specific manifests
ksail workload apply
```

### Secrets Management

Encrypt and decrypt sensitive configuration files with SOPS:

```bash
# Encrypt a file
ksail cipher encrypt secrets.yaml

# Decrypt a file
ksail cipher decrypt secrets.enc.yaml
```

### Configuration

KSail uses declarative configuration files:

- `ksail.yaml` - Main KSail configuration
- `kind.yaml` / `k3d.yaml` - Distribution-specific cluster config
- `k8s/` - Directory containing Kubernetes manifests

Example `ksail.yaml`:

```yaml
apiVersion: ksail.devantler.com/v1alpha1
kind: KSailCluster
metadata:
  name: my-cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
```

## Documentation

### Command Reference

Get detailed help for any command:

```bash
ksail --help                    # Main help
ksail init --help              # Project initialization
ksail cluster --help           # Cluster management
ksail workload --help          # Workload operations
ksail cipher --help            # Secrets management
```

### Project Structure

After running `ksail init`, your project will have:

```
my-project/
├── ksail.yaml              # KSail configuration
├── kind.yaml               # Cluster distribution config
└── k8s/                    # Kubernetes manifests directory
    ├── namespace.yaml
    ├── deployment.yaml
    └── service.yaml
```

### Package Documentation

Explore the codebase and API documentation:

- **[cmd/](./cmd/README.md)** - CLI command implementations
- **[pkg/](./pkg/README.md)** - Core business logic and public APIs
- **[API Reference](./pkg/apis/cluster/v1alpha1/README.md)** - Custom resource definitions

### Examples

Check out example projects and configurations:

- [Example Kind Cluster](https://github.com/devantler-tech/ksail-go/tree/main/examples/kind) (Coming soon)
- [Example K3d Cluster](https://github.com/devantler-tech/ksail-go/tree/main/examples/k3d) (Coming soon)
- [Multi-Cluster Setup](https://github.com/devantler-tech/ksail-go/tree/main/examples/multi-cluster) (Coming soon)

### Contributing

Interested in contributing? See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on:

- Development setup
- Running tests
- Code style and linting
- Submitting pull requests

## Structure

The KSail Go project is organized into several main packages:

- **[cmd/](./cmd/README.md)** - Command-line interface implementation using Cobra framework
- **[pkg/](./pkg/)** - Core business logic packages:
  - **[pkg/apis/](./pkg/apis/cluster/v1alpha1/README.md)** - Kubernetes API definitions
  - **[pkg/config-manager/](./pkg/config-manager/README.md)** - Configuration management utilities
  - **[pkg/installer/](./pkg/installer/README.md)** - Component installation utilities
  - **[pkg/io/](./pkg/io/README.md)** - Safe file I/O operations with security features and path helpers
  - **[pkg/provisioner/](./pkg/provisioner/README.md)** - Cluster provisioning and lifecycle management
- **internal/** - Internal utility packages:
  - **[internal/testutils/](./internal/testutils/README.md)** - Shared testing utilities

Each package contains detailed documentation about its purpose, features, and usage examples.

## Related Projects

KSail builds upon and integrates with these excellent Kubernetes tools:

- **[Kind](https://kind.sigs.k8s.io/)** - Kubernetes in Docker, used for local cluster provisioning
- **[K3d](https://k3d.io/)** - K3s in Docker, lightweight Kubernetes distribution
- **[eksctl](https://eksctl.io/)** - Official CLI for Amazon EKS
- **[k9s](https://k9scli.io/)** - Terminal-based UI for Kubernetes cluster management
- **[SOPS](https://github.com/getsops/sops)** - Secrets management and encryption
- **[Flux CD](https://fluxcd.io/)** - GitOps toolkit for Kubernetes (planned integration)

### KSail Ecosystem

- **[KSail (Original)](https://github.com/devantler-tech/ksail)** - The original .NET implementation of KSail

## Presentations

Stay tuned for upcoming presentations, tutorials, and demos!

## Star History

---

Contributions welcome. Open an issue or PR to propose features.
