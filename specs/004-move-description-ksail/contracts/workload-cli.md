# CLI Contract: Workload Command Group

## Overview

Defines the expected interface for the new `ksail workload` command group, including user-visible help text fragments, outputs, and exit codes for each subcommand.

## Command Matrix

| Command | Description | Output | Exit Code |
|---------|-------------|--------|-----------|
| `ksail workload` | Parent namespace summarizing workload operations. | Help text lists `reconcile`, `apply`, `install` subcommands. | 0 |
| `ksail workload reconcile` | Placeholder reconciliation command. | `notify.Infoln` message: "Workload reconciliation coming soon." | 0 |
| `ksail workload apply` | Placeholder manifest apply command. | `notify.Infoln` message: "Workload apply coming soon." | 0 |
| `ksail workload install` | Placeholder Helm install command. | `notify.Infoln` message: "Workload install coming soon." | 0 |

## Help Text Contract

- `ksail workload --help` MUST include:
  - Short description: "Manage workload operations"
  - Usage block with `ksail workload [command]`
  - Subcommand listing for `reconcile`, `apply`, `install`
- Each subcommand `--help` MUST describe future intent and indicate current placeholder status.

## Error Contract

- Invoking `ksail reconcile` MUST return exit code 1 with stderr containing:
  - Cobra's standard unknown command prefix.
  - Additional guidance: `Command "reconcile" moved to "ksail workload reconcile".`
- Unknown commands under `ksail workload` continue to rely on Cobra's default suggestion behavior.

## Testing Hooks

- Snapshot tests capture `ksail workload --help` and subcommand help outputs.
- Unit tests assert placeholder commands write the correct strings and exit successfully.
- Error handling test ensures guidance appended when legacy command is invoked.
