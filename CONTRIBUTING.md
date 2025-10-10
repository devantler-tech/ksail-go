# Contributing

> [!NOTE]
> If you are looking to become a maintainer of ksail, please reach out to @devantler on <https://devantler.tech/contact/>. I can facilitate introductions, mentorship, and regular check-ins, while also providing premium GitHub features to maintainers. As a maintainer, you will have the opportunity to obtain mandate to steer the project's direction in collaboration with me and the community.

This project accepts contributions in the form of [**bug reports**](https://github.com/devantler-tech/ksail-go/issues/new/choose), [**feature requests**](https://github.com/devantler-tech/ksail-go/issues/new/choose), and **pull requests** (see below). If you are looking to contribute code, please follow the guidelines outlined in this document.

## Getting Started

To get started with contributing to ksail, you'll need to set up your development environment, and understand the codebase, the CI setup and its requirements.

To understand the codebase it is recommended to read the `.github/copilot-instructions.md` file, which provides an overview of the project structure and key components. You can also use GitHub Copilot to assist you in navigating the codebase and understanding its functionality.

## Project Structure

The KSail Go project is organized into several main directories:

### Command-Line Interface

- **[cmd/](./cmd/README.md)** - CLI implementation using Cobra framework:
  - **cmd/cipher/** - SOPS cipher management commands
  - **cmd/cluster/** - Cluster lifecycle commands (create, delete, start, stop, connect, info, list)
  - **cmd/workload/** - Workload management commands (logs, etc.)
  - **[cmd/internal/](./cmd/internal/README.md)** - Internal command utilities and helpers

### Core Packages

- **[pkg/](./pkg/)** - Core business logic packages:
  - **[pkg/apis/](./pkg/apis/cluster/v1alpha1/README.md)** - Kubernetes API definitions and custom resource types
  - **[pkg/config-manager/](./pkg/config-manager/README.md)** - Configuration management for KSail and distribution configs
  - **[pkg/containerengine/](./pkg/containerengine/README.md)** - Container engine abstraction (Docker, Podman)
  - **pkg/di/** - Dependency injection helpers for commands
  - **pkg/errorhandler/** - Centralized error handling and formatting
  - **pkg/helm/** - Helm client implementation
  - **[pkg/installer/](./pkg/installer/README.md)** - Component installation utilities (kubectl, Flux, etc.)
  - **[pkg/io/](./pkg/io/README.md)** - Safe file I/O operations with security features
  - **[pkg/k9s/](./pkg/k9s/README.md)** - k9s terminal UI integration
  - **pkg/kubectl/** - kubectl client implementation
  - **[pkg/provisioner/](./pkg/provisioner/README.md)** - Cluster provisioning and lifecycle management
  - **[pkg/scaffolder/](./pkg/scaffolder/README.md)** - Project scaffolding and file generation
  - **pkg/sops/** - SOPS encryption client implementation
  - **[pkg/ui/](./pkg/ui/README.md)** - User interface utilities (notifications, ASCII art, timing)
  - **[pkg/validator/](./pkg/validator/README.md)** - Configuration validation utilities

### Internal Packages

- **internal/** - Internal utility packages:
  - **[internal/testutils/](./internal/testutils/README.md)** - Shared testing utilities and helpers

Each package contains detailed documentation about its purpose, features, and usage examples. Packages in `pkg/` are part of KSail's public API, while packages in `internal/` and `cmd/internal/` are for internal use only.

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go (v1.23.9+)](https://go.dev/doc/install)
- [mockery](https://vektra.github.io/mockery/v3.5/installation/)
- [golang-ci](https://golangci-lint.run/docs/welcome/install/)
- [mega-linter](https://megalinter.io/latest/mega-linter-runner/#installation)
- [Docker](https://www.docker.com/get-started/)

### Lint

KSail uses mega-linter with the go flavor, and uses a strict configuration to ensure code quality and consistency. You can run the linter with the following command:

```sh
# working-directory: ./
mega-linter-runner -f go
```

The same configuration is used in CI, so you can expect the same linting behavior in your local environment as in the CI pipeline.

### Build

```sh
# working-directory: ./
go build ./...
```

### Test

#### Generating mocks

```sh
# working-directory: ./
mockery
```

#### Unit tests

```sh
# working-directory: ./
go test ./...
```

## CI

### Pre-commit Hooks

> **Note**: Pre-commit hooks are automatically executed for user pushes through the [pre-commit.ci](https://pre-commit.ci/) GitHub app, which validates and runs these hooks if you forget to configure them locally or push without hooks enabled. This automatic execution only applies to user pushes and not bot pushes.

KSail uses pre-commit hooks to ensure code quality and consistency before commits are made. This is done via the [pre-commit framework](https://pre-commit.com/). Active hooks are defined in the `.pre-commit-config.yaml` file.

To use these hooks, install pre-commit and run:

```sh
pre-commit install
```

### GitHub Workflows

#### Unit Tests

```sh
# working-directory: ./
go test ./...
```

#### System Tests

System tests are configured in a GitHub Actions workflow file located at `.github/workflows/ci.yaml`. These test e2e scenarios for various providers and configurations. You are unable to run these tests locally, but they are required in CI, so breaking changes will result in failed checks.

## CD

### Release Process

The release process for KSail is fully-automated and relies on semantic versioning. When PRs are merged into the main branch, a new version is automatically released based on the name of the PR. The following conventions are used:

- **fix:** A patch release (e.g. 1.0.1) is triggered.
- **feat:** A minor release (e.g. 1.1.0) is triggered.
- **BREAKING CHANGE:** A major release (e.g. 2.0.0) is triggered.

The changelog is auto-generated by go-releaser, so contributors just have to ensure their PRs are well-named and descriptive, such that the intent of the changes is clear.
