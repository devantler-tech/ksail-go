# Data Model – KSail Cluster Command Hierarchy

## Command Tree

| Parent Command | Subcommand | Purpose | Notes |
|----------------|------------|---------|-------|
| `ksail`        | `cluster`  | Parent namespace for cluster lifecycle flows | Root help must list this command with concise description |
| `ksail cluster` | `up` | Start or create the target cluster using existing configuration | Reuses handler `HandleUpRunE`; inherits distribution/context flags |
| `ksail cluster` | `down` | Tear down cluster resources | Reuses `HandleDownRunE` |
| `ksail cluster` | `start` | Start a previously created but stopped cluster | Reuses `HandleStartRunE` |
| `ksail cluster` | `stop` | Stop a running cluster | Reuses `HandleStopRunE` |
| `ksail cluster` | `status` | Show current cluster state | Reuses `HandleStatusRunE` |
| `ksail cluster` | `list` | Enumerate available clusters | Reuses `HandleListRunE` |
<!-- Note: `reconcile` is intentionally excluded from `ksail cluster` in this refactor. It will be migrated to `ksail workloads reconcile` later. -->

## Flags & Shared Options

| Flag Scope | Flag | Source Helper | Notes |
|------------|------|---------------|-------|
| All cluster subcommands | Distribution selectors | `cmdhelpers.StandardDistributionFieldSelector` | Already attached to each command constructor; ensure migration preserves wiring |
| All cluster subcommands | Execution context selectors | `cmdhelpers.StandardContextFieldSelector` | No changes, but verify they work when nested |
| `cluster up` | Timeout | `configmanager.FieldSelector[v1alpha1.Cluster]` | Default remains `5m`; ensure help text surfaces under new parent |

## Output Contracts

| Command | Success Output (stdout) | Error Handling |
|---------|-------------------------|----------------|
| `cluster up` | "Cluster created and started successfully (stub implementation)" | Wrap errors with `failed to handle cluster command` as today |
| `cluster down` | Matches existing stub message | Same error wrapping |
| Others | No change from existing implementations | Propagate existing error formatting |

## State Transitions

| Command | Previous State | Next State | Notes |
|---------|----------------|------------|-------|
| `cluster up` | Unprovisioned / stopped | Running | Stub implementation today; behavior unchanged |
| `cluster down` | Any | Removed | Stub implementation |
| `cluster start` | Stopped | Running | |
| `cluster stop` | Running | Stopped | |

## Documentation Touchpoints

- Root command help must list `cluster` with concise description (clarification #3).
- `cluster` command help should enumerate subcommands with short/long descriptions derived from existing command metadata.
