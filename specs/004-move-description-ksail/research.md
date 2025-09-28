# Phase 0 Research Summary

## Command Namespace & Structure

- **Decision**: Introduce a singular `workload` command group implemented in a dedicated `cmd/workload` package that registers subcommands for `reconcile`, `apply`, and `install`.
- **Rationale**: Keeps naming aligned with the spec clarification while matching existing folder conventions (e.g., `cmd/cluster`). Centralizing the group simplifies future growth.
- **Alternatives Considered**:
  - Retain a plural namespace (`workloads`): rejected per clarification.
  - Co-locate commands within `cmd/root.go`: rejected to avoid bloating the root file and to follow existing modular command structure.

## Placeholder Command Behavior

- **Decision**: Each workload subcommand will emit a standardized "Coming soon" notice via `notify.Infoln` and return `nil` so the CLI exits with code 0.
- **Rationale**: Satisfies requirements for communicating future work without failing the command, maintains consistent UX, and leverages existing notification utilities.
- **Alternatives Considered**:
  - Returning an error until functionality ships: rejected because it would violate the requirement to exit successfully while communicating intent.
  - Printing directly with `fmt.Println`: rejected to maintain the unified notify system used across the CLI.

## Legacy Command Messaging

- **Decision**: Remove the top-level `reconcile` command and intercept unknown-command errors inside `runWithArgs` to append a guidance message that points to `ksail workload reconcile` when the removed command is used.
- **Rationale**: Keeps Cobra wiring simple while ensuring the user sees a targeted migration hint. Handling it in `runWithArgs` avoids deep changes to Cobra internals.
- **Alternatives Considered**:
  - Leaving a hidden alias for `reconcile`: rejected because it conflicts with the requirement to remove the legacy command.
  - Adding a custom `Command` just to print the guidance: rejected because it would still expose the command and contradict the removal requirement.

## Testing Approach

- **Decision**: Extend existing snapshot-based command tests (using `go-snaps`) for help text, and add explicit unit tests asserting the placeholder outputs and success exit codes.
- **Rationale**: Aligns with the constitution's TDD requirement and matches existing testing patterns in `cmd`.
- **Alternatives Considered**:
  - Relying solely on manual QA: rejected per constitution (tests are mandatory).
  - Writing only integration tests: rejected because unit-level coverage is necessary to verify error routing and placeholder behavior.
