# Feature Specification: Timing Output Control

**Feature Branch**: `001-timing-output-control`
**Created**: 2025-12-12
**Status**: Draft
**Input**: Issue #601: <https://github.com/devantler-tech/ksail-go/issues/601>

## Clarifications

### Session 2025-12-12

- Q: When timing is enabled, when should the timing block be printed? → A: After each timed activity (so `current` is that activity; `total` accumulates across activities)
- Q: Should timing be configurable via `ksail.yaml`? → A: No, timing is controlled only via a CLI flag.
- Q: What counts as a “timed activity” for printing timing output? → A: Each progress/spinner step (each unit that emits a `✔ completion message`).

## Constitution Constraints _(mandatory)_

These are NON-NEGOTIABLE constraints derived from `.specify/memory/constitution.md`.

- **Interface-first**: Define interfaces before implementations; design for mocking.
- **Test-first**: Plan and write tests first; tests cover public APIs only (no white-box).
- **Package-first**: Business logic lives in `pkg/`; `cmd/` remains a thin wrapper.
- **Quality gates**: Work must pass `mockery`, `go test ./...`, `golangci-lint run`, `go build ./...`.
- **KISS/DRY/YAGNI**: Prefer simple, non-duplicated, non-speculative solutions.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Enable timing for a single run (Priority: P1)

As a CLI user, I want to enable timing output for a single command run using a `--timing` option, so that timing information is only shown when I explicitly request it.

**Why this priority**: This is the fastest, least surprising way to opt into timing without changing any persistent settings.

**Independent Test**: Run any command twice—once without `--timing` and once with `--timing`—and verify only the second run includes the timing block.

**Acceptance Scenarios**:

1. **Given** the user does not pass `--timing`, **When** a user runs a command, **Then** the output does not include any timing block.
2. **Given** a user runs a command with `--timing`, **When** a timed activity completes, **Then** the output includes a timing block formatted as specified.

---

### User Story 2 - Consistent timing output format (Priority: P3)

As a CLI user, I want timing output to be consistently formatted, so that CLI output remains readable and predictable.

**Why this priority**: A stable, readable format improves usability and reduces noise in logs/screenshots.

**Independent Test**: Enable timing and verify the output matches the documented format exactly.

**Acceptance Scenarios**:

1. **Given** timing output is enabled, **When** a timed activity completes, **Then** output includes:

- a completion message line
- followed by a timing block with `current` and `total` durations

2. **Given** timing output is enabled for a command that performs multiple timed activities, **When** each activity completes, **Then** `total` reflects the accumulated duration of all activities completed so far.

---

### Edge Cases

- If timing is enabled but a duration cannot be determined for an activity, the output still renders and uses a clearly defined zero/empty duration value.
- Output formatting remains stable across different commands (consistent labels and indentation).

## Assumptions

- The `--timing` option is available in a way that applies consistently to user-invoked commands.
- Durations are rendered in a human-readable duration format consistent across the CLI.

## Dependencies

- Existing command execution already emits a completion message suitable for appending timing output.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The CLI MUST support a boolean `--timing` option that enables timing output for the current invocation.
- **FR-002**: Timing output MUST be OFF by default unless enabled by the `--timing` flag.
- **FR-003**: When timing output is enabled, the output MUST include the following block (including labels and indentation):

  ```
   ✔ completion message
   ⏲ current: <duration>
     total:  <duration>
  ```

- **FR-004**: When timing output is enabled, the timing block MUST be printed after each timed activity completes.

- **FR-005**: `current` MUST represent the duration of the most recently completed activity; `total` MUST represent the accumulated duration of all activities completed so far within the invocation.
- **FR-006**: Documentation MUST describe the `--timing` option, the default behavior, and the output format.
- **FR-007**: Automated tests MUST cover the flag activation point and MUST validate behavior via public APIs only (no white-box tests).

### Key Entities _(include if feature involves data)_

- **Timing setting**: A boolean user preference that can be provided via CLI option.
- **Timed activity**: A single progress/spinner step (a unit that emits a `✔ completion message`) that can produce a duration.
- **Timing output block**: A user-visible text block containing `current` and `total` durations.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: With default settings, timing output is not shown in command output.
- **SC-002**: With timing enabled via `--timing`, the timing output block matches the documented format exactly.
- **SC-002a**: With timing enabled via `--timing`, each `✔ completion message` is immediately followed by a timing block where `current` is that step and `total` accumulates across steps.
- **SC-003**: Automated tests verify flag-based activation using only public interfaces (CLI behavior) and pass in CI.
