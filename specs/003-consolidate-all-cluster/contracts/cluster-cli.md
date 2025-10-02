# Contract – KSail `cluster` Command Surface

## Command Summary

| Command | Description | Expected Exit Code |
|---------|-------------|--------------------|
| `ksail cluster` | Parent command providing cluster lifecycle actions; shows help when run without subcommand | 0 on help display |
| `ksail cluster up` | Provision or start cluster according to project configuration | 0 on success, non-zero on failure |
| `ksail cluster down` | Destroy cluster resources | 0/!0 |
| `ksail cluster start` | Start an existing cluster | 0/!0 |
| `ksail cluster stop` | Stop a running cluster | 0/!0 |
| `ksail cluster status` | Report cluster status | 0/!0 |
| `ksail cluster list` | List managed clusters | 0/!0 |

> **Note:** `ksail cluster reconcile` is intentionally excluded from this refactor. It will be migrated to `ksail workloads reconcile` later.

## Help Output Requirements

- `ksail --help` **must** list `cluster` alongside other top-level commands with a short description: "Manage cluster lifecycle" (exact phrasing finalized during implementation but must be concise).
- `ksail cluster --help` **must** include: short description, long description referencing lifecycle operations, and list each subcommand with its short description.

## Flag Expectations

- All subcommands inherit standard distribution and context flags by delegating to existing helper selectors.
- `ksail cluster up` exposes the timeout option with default `5m` via `--timeout` from `configmanager.FieldSelector[v1alpha1.Cluster]`.
- No new flags introduced; any renamed flags would be a breaking change and require constitutional review.

## Error Handling Contract

- Legacy commands (`ksail up`, etc.) must not exist; invoking them should yield Cobra’s default "unknown command" error.
- Subcommand execution should continue to wrap errors using existing helper `cmdhelpers.HandleSimpleClusterCommand` patterns to maintain consistency with tests.

## Testing Checklist

- Unit tests verify each subcommand’s metadata (use, short, long) and ensure they are registered beneath `cluster`.
- Root command test asserts `cluster` is present in the available subcommand list.
- Snapshot tests (`cmd/__snapshots__`) updated if necessary to reflect new help output.
