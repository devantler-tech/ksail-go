[![codecov](https://codecov.io/gh/devantler-tech/ksail-go/graph/badge.svg?token=HSUfhaiXwq)](https://codecov.io/gh/devantler-tech/ksail-go)
[![CI - Go](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml)

# ksail

> [!IMPORTANT]
> This is a work in progress to migrate KSail to a Golang. This is a huge endeavour, but being able to leverage the power of the Go ecosystem will be invaluable. The amount of packages available in Go to support this project is immense, so switching programming language has the potential to greatly enhance the functionality, performance and ease of use of KSail. I also hope switching will promote adoption and contributions.

Take control of Kubernetes without the chaos. ⚡ KSail is your all-in-one SDK for spinning up clusters and managing workloads—right from your own machine. Instead of juggling a dozen CLI tools, KSail streamlines your workflow with a single, declarative interface built on top of the Kubernetes tools you already know and trust.

🌟 Declarative. Local. Effortless. Welcome to Kubernetes, simplified.

## Getting Started

### Prerequisites

- Linux (amd64 and arm64)
- MacOS (amd64 and arm64)

### Installation

#### Homebrew

#### Go Install

```bash
go install github.com/devantler-tech/ksail-go@latest
ksail --help
```

#### From Source

```bash
git clone https://github.com/devantler-tech/ksail-go.git
go build -o ksail
./ksail --help
```

## Usage

Run the CLI:

```bash
ksail --help
```

## Documentation

## Structure

The KSail Go project is organized into several main packages:

- **[cmd/](./cmd/README.md)** - Command-line interface implementation using Cobra framework
- **[pkg/](./pkg/)** - Core business logic packages:
  - **[pkg/apis/](./pkg/apis/cluster/v1alpha1/README.md)** - Kubernetes API definitions
  - **[pkg/config-manager/](./pkg/config-manager/README.md)** - Configuration management utilities
  - **[pkg/installer/](./pkg/installer/README.md)** - Component installation utilities
  - **[pkg/io/](./pkg/io/README.md)** - Safe file I/O operations with security features
  - **[pkg/provisioner/](./pkg/provisioner/README.md)** - Cluster provisioning and lifecycle management
- **[internal/](./internal/)** - Internal utility packages:
  - **[internal/utils/k8s/](./internal/utils/k8s/README.md)** - Kubernetes utilities
  - **[internal/utils/path/](./internal/utils/path/README.md)** - Path utilities

Each package contains detailed documentation about its purpose, features, and usage examples.

## Related Projects

## Presentations

## Star History

---

Contributions welcome. Open an issue or PR to propose features.
