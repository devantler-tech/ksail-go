[![Go Reference](https://pkg.go.dev/badge/github.com/devantler-tech/ksail-go.svg)](https://pkg.go.dev/github.com/devantler-tech/ksail-go)
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
cd ksail-go
go build -o ksail
./ksail --help
```

## Usage ⚙️

Get a Kubernetes cluster running in seconds:

```bash
# Initialize a new project with Kind
ksail cluster init --distribution Kind

# Create and start the cluster
ksail cluster create

# Deploy your workloads
ksail workload reconcile

# Clean up when done
ksail cluster delete
```

## Documentation 📚

### For Users 📖

- Browse the documentation on [`devantler-tech/ksail-docs`](https://github.com/devantler-tech/ksail-docs) (Markdown) or on <https://ksail.devantler.tech> (GitHub Pages).
- Run `ksail --help` or `ksail <command> --help` for the latest CLI flags.

### For Contributors 👥

- [CONTRIBUTING.md](./CONTRIBUTING.md) — Contribution guide, development prerequisites, and workflows
- [.github/copilot-instructions.md](./.github/copilot-instructions.md) — GitHub Copilot configuration and best practices
- [API Documentation](https://pkg.go.dev/github.com/devantler-tech/ksail-go) — Go package documentation

## Flux Installer ☸️

Flux installation during cluster bootstrap is handled by `pkg/svc/installer/flux`, a Helm-based implementation of the shared `pkg/svc/installer` interface. This installer provisions or upgrades the Flux Operator whenever `spec.gitOpsEngine: Flux` is configured, keeping the bootstrap experience consistent with other component installers.

Because Flux is handled automatically when configured, the `ksail cluster flux` command has been removed.

## Related Projects 🔗

## Presentations 🎤

## Star History ⭐
