# Quickstart: Workload Command Restructure

## Prerequisites

- Go 1.24.0+ installed (`go version`)
- Project dependencies downloaded (`go mod download`)
- Snapshot tooling available via `go-snaps` (already in go.mod)

## Steps

1. **Run unit tests first (TDD expectation)**

   ```bash
   cd /Users/ndam/git-personal/monorepo/projects/ksail-go
   go test ./cmd/... ./internal/... ./pkg/...
   ```

2. **Build the CLI locally**

   ```bash
   cd /Users/ndam/git-personal/monorepo/projects/ksail-go
   go build -o ksail .
   ```

3. **Verify new command namespace help**

   ```bash
   ./ksail workload --help
   ```

   - Confirms the `workload` group is discoverable and lists subcommands.
4. **Check placeholder outputs**

   ```bash
   ./ksail workload reconcile
   ./ksail workload apply
   ./ksail workload install
   ```

   - Each command should print the informational message (`ℹ Workload <action> coming soon.`) and exit with status 0.
5. **Observe legacy command guidance**

   ```bash
   ./ksail reconcile
   ```

   - CLI should report the command is unknown and direct the user to `ksail workload reconcile`.
6. **Review help text snapshots**

   ```bash
   go test ./cmd -run TestWorkloadHelp -update
   ```

   - (Optional) Refresh snapshots after intentional help text edits.

## Cleanup

- Remove local binary if desired: `rm ./ksail`
