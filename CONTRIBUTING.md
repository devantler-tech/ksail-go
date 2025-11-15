# Contributing

> [!NOTE]
> If you are looking to become a maintainer of ksail, please reach out to @devantler on <https://devantler.tech/contact/>. I can facilitate introductions, mentorship, and regular check-ins, while also providing premium GitHub features to maintainers. As a maintainer, you will have the opportunity to obtain mandate to steer the project's direction in collaboration with me and the community.

This project accepts contributions in the form of [**bug reports**](https://github.com/devantler-tech/ksail-go/issues/new/choose), [**feature requests**](https://github.com/devantler-tech/ksail-go/issues/new/choose), and **pull requests** (see below). If you are looking to contribute code, please follow the guidelines outlined in this document.

## Getting Started

To get started with contributing to ksail, you'll need to set up your development environment, and understand the codebase, the CI setup and its requirements.

To understand the codebase it is recommended to read the `.github/copilot-instructions.md` file, which provides an overview of the project structure and key components. You can also use GitHub Copilot to assist you in navigating the codebase and understanding its functionality.

### Code Documentation

For detailed package and API documentation, refer to the Go documentation at [pkg.go.dev/github.com/devantler-tech/ksail-go](https://pkg.go.dev/github.com/devantler-tech/ksail-go). This provides comprehensive documentation for all exported packages, types, functions, and methods.

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

## Project Structure

The repository is organized into three main directories:

- **cmd/** - CLI command implementations
- **pkg/** - Public packages (importable by external projects)
- **internal/** - Internal packages (not importable externally)

For detailed package and API documentation, refer to [pkg.go.dev/github.com/devantler-tech/ksail-go](https://pkg.go.dev/github.com/devantler-tech/ksail-go).

## Adding New CNI Installers

CNI (Container Network Interface) installers are located under `pkg/svc/installer/cni/`. To add support for a new CNI:

1. **Create a new subdirectory** under `pkg/svc/installer/cni/` (e.g., `pkg/svc/installer/cni/mycni/`)

2. **Implement the installer.Installer interface** in your new package:
   ```go
   package mycniinstaller
   
   import (
       "context"
       "fmt"
       "time"
       "github.com/devantler-tech/ksail-go/pkg/client/helm"
       "github.com/devantler-tech/ksail-go/pkg/k8s"
       "github.com/devantler-tech/ksail-go/pkg/svc/installer"
       "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
   )
   
   // MyCNIInstaller implements installer.Installer by embedding InstallerBase.
   type MyCNIInstaller struct {
       *cni.InstallerBase
   }
   
   // NewMyCNIInstaller creates a new MyCNI installer instance.
   func NewMyCNIInstaller(
       client helm.Interface,
       kubeconfig, context string,
       timeout time.Duration,
   ) *MyCNIInstaller {
       mycniInstaller := &MyCNIInstaller{}
       mycniInstaller.InstallerBase = cni.NewInstallerBase(
           client,
           kubeconfig,
           context,
           timeout,
           mycniInstaller.waitForReadiness,
       )
       return mycniInstaller
   }
   
   // Install installs the CNI using Helm and waits for readiness.
   func (i *MyCNIInstaller) Install(ctx context.Context) error {
       client, err := i.GetClient()
       if err != nil {
           return fmt.Errorf("get helm client: %w", err)
       }
   
       // Configure Helm repository
       repoConfig := helm.RepoConfig{
           Name:     "mycni",
           URL:      "https://helm.mycni.io",
           RepoName: "mycni",
       }
   
       // Configure Helm chart installation
       chartConfig := helm.ChartConfig{
           ReleaseName:     "mycni",
           ChartName:       "mycni/mycni",
           Namespace:       "kube-system",
           RepoURL:         "https://helm.mycni.io",
           CreateNamespace: false,
           SetJSONVals:     map[string]string{"replicas": "1"},
       }
   
       // Install or upgrade the chart
       err = helm.InstallOrUpgradeChart(ctx, client, repoConfig, chartConfig, i.GetTimeout())
       if err != nil {
           return fmt.Errorf("install or upgrade mycni: %w", err)
       }
   
       return nil
   }
   
   // waitForReadiness waits for MyCNI pods to be ready.
   func (i *MyCNIInstaller) waitForReadiness(ctx context.Context) error {
       checks := []k8s.ReadinessCheck{
           {Type: "daemonset", Namespace: "kube-system", Name: "mycni"},
       }
   
       err := installer.WaitForResourceReadiness(
           ctx,
           i.GetKubeconfig(),
           i.GetContext(),
           checks,
           i.GetTimeout(),
           "mycni",
       )
       if err != nil {
           return fmt.Errorf("wait for mycni readiness: %w", err)
       }
   
       return nil
   }
   ```
   
   For a complete implementation pattern, see:
   - `pkg/svc/installer/cni/cilium/installer.go`
   - `pkg/svc/installer/cni/calico/installer.go`

3. **Reuse shared utilities** from `cni.InstallerBase` for Helm chart installation and readiness checks

4. **Add comprehensive unit tests** following patterns in existing CNI implementations (see `pkg/svc/installer/cni/cilium/` or `pkg/svc/installer/cni/calico/`)

5. **Update documentation** to reflect the new CNI option

For detailed guidance and code examples, see `specs/001-cni-installer-move/quickstart.md`.

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
