[![codecov](https://codecov.io/gh/devantler-tech/ksail-go/graph/badge.svg?token=HSUfhaiXwq)](https://codecov.io/gh/devantler-tech/ksail-go)
[![CI - Go](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml)

# ksail

> [!IMPORTANT]
> This is a work in progress to migrate KSail to a Golang. This is a huge endeavour, but being able to leverage the power of the Go ecosystem will be invaluable. The amount of packages available in Go to support this project is immense, so switching programming language has the potential to greatly enhance the functionality, performance and ease of use of KSail. I also hope switching will promote adoption and contributions.

**What is KSail?** 🚢

KSail is a CLI tool for managing local Kubernetes clusters and workloads with a single, unified interface. It wraps popular tools like Kind and K3d, providing a consistent experience across different distributions—no more juggling multiple CLIs or memorizing different syntaxes.

**Why KSail?** 💡

Setting up local Kubernetes environments typically means learning multiple tools, each with their own commands and configuration formats. KSail eliminates this friction by providing:

- 🎯 A single command-line interface for Kind and K3d clusters
- 📝 Declarative configuration for reproducible environments
- 🔐 Integrated workload and secrets management
- ⚡ Fast cluster lifecycle operations (create, start, stop, delete)

Whether you're developing applications, testing infrastructure changes, or learning Kubernetes, KSail gets you from zero to a working cluster in seconds.

🌟 Declarative. Local. Effortless. Welcome to Kubernetes, simplified.

## Getting Started 🚀

### Prerequisites ✅

**System Requirements:**

- 🐧 Linux (amd64 and arm64)
- 🍎 MacOS (amd64 and arm64)
- 🐳 Docker (required for Kind and K3d clusters)

### Installation 📦

#### Homebrew 🍺

#### Go Install 🔧

```bash
go install github.com/devantler-tech/ksail-go@latest
ksail --help
```

#### From Source 💻

```bash
git clone https://github.com/devantler-tech/ksail-go.git
go build -o ksail
./ksail --help
```

## Usage ⚙️

Get a Kubernetes cluster running in seconds:

```bash
# Initialize a new project with Kind
ksail init --distribution Kind

# Create and start the cluster
ksail cluster create

# Deploy your workloads
ksail workload reconcile

# Clean up when done
ksail cluster delete
```

For detailed command reference, run `ksail --help`.

## Documentation 📚

For information on how to contribute, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Structure 🏗️

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

## Related Projects 🔗

## Presentations 🎤

## Star History ⭐

---

Contributions welcome. Open an issue or PR to propose features.
