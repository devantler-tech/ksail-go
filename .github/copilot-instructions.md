# KSail-Go – Copilot Instructions

## Quick Start (local)

- Go toolchain: `go 1.25.4` (see `go.mod`; CI uses `actions/setup-go` with `go-version-file`).
- Build/test: `go build ./...`, `go test ./...`.
- Lint/format: `golangci-lint run --timeout 5m --fix` (see `.golangci.yml`; strict `depguard` allowlist).
- Mocks: run `mockery` (see `.mockery.yml`; generates `mocks.go` in package dirs—don’t hand-edit).
- Schema: `./.github/scripts/generate-schema.sh` regenerates `schemas/ksail-config.schema.json` (CI auto-commits changes).

## Big Picture

- CLI is Cobra: `main.go` → `cmd.NewRootCmd()` → subcommands in `cmd/cluster`, `cmd/workload`, `cmd/cipher`.
- Error UX: Cobra commands typically set `SilenceUsage: true`; root execution is wrapped by `pkg/ui/error-handler` and user-facing messages use `pkg/ui/notify`.

## Dependency Injection (how commands are wired)

- `pkg/di` wraps `samber/do` to keep commands testable; root uses `runtime.NewRuntime()`.
- Prefer `runtime.RunEWithRuntime(runtimeContainer, ...)` and decorators like `runtime.WithTimer(...)` over global state.
- Default providers live in `pkg/di/providers.go` (e.g. `timer.Timer`, `clusterprovisioner.Factory`).

## Config Model (ksail.yaml + env + flags)

- Config is `ksail.yaml` loaded via Viper (`pkg/io/config-manager/ksail`). Priority: `defaults < file < env < flags`.
- Config discovery walks up parent directories to find `ksail.yaml` (like Git) and also checks standard paths (see `pkg/io/config-manager/ksail/viper.go`).
- Env vars use prefix `KSAIL_`; `.` and `-` are mapped to `_` (tests in `pkg/io/config-manager/ksail/viper_test.go`).
- Commands typically construct a `ksailconfigmanager.ConfigManager` with field selectors and call `LoadConfig()` / `LoadConfigSilent()`.

## Cluster Lifecycle & Installers

- `cmd/cluster` orchestrates lifecycle and calls provisioners under `pkg/svc/provisioner/*`.
- Component installs are `installer.Installer` implementations in `pkg/svc/installer/*`.
  - CNIs live under `pkg/svc/installer/cni/*` and embed `cni.InstallerBase`.
  - Flux uses an OCI Helm chart installer in `pkg/svc/installer/flux`.

## Workload Commands

- Many `cmd/workload/*` commands are thin wrappers over clients in `pkg/client/*` (kubectl/flux). They often resolve kubeconfig via `pkg/cmd.GetKubeconfigPathSilently()`.

## Tests

- Unit tests use `testify` + generated mocks (see `.mockery.yml` and patterns like `helm.NewMockInterface(t)` in installer tests).
- Snapshot tests exist under `cmd/__snapshots__/` (via `go-snaps`).

## Active Technologies
- Go 1.25.4 + Cobra, Viper, samber/do (DI), fatih/color, go-snaps, testify (001-timing-output-control)

## Recent Changes
- 001-timing-output-control: Added Go 1.25.4 + Cobra, Viper, samber/do (DI), fatih/color, go-snaps, testify
