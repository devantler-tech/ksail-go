[![codecov](https://codecov.io/gh/devantler-tech/ksail-go/graph/badge.svg?token=HSUfhaiXwq)](https://codecov.io/gh/devantler-tech/ksail-go)
[![CI - Go](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/devantler-tech/ksail-go/actions/workflows/ci.yaml)

# ksail

> [!IMPORTANT]
> This is a work in progress to migrate KSail to a Golang. This is a huge endeavour, but being able to leverage the power of the Go ecosystem will be invaluable. The amount of packages available in Go to support this project is immense, so switching programming language has the potential to greatly enhance the functionality, performance and ease of use of KSail. I also hope switching will promote adoption and contributions.

KSail is a CLI tool with the ambition to become a full-fledged SDK for creating and maintaining Kubernetes clusters—locally or in the cloud. It provides a unified interface for managing clusters and workloads across different distributions (currently Kind and K3d, with more planned). By wrapping existing tools with a consistent command-line experience, KSail eliminates the complexity of juggling multiple CLIs and learning different syntaxes for each distribution.

KSail simplifies your Kubernetes workflow by providing:

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

## Related Projects 🔗

## Presentations 🎤

## Star History ⭐

---

Contributions welcome. Open an issue or PR to propose features.
