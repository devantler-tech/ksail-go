# Feature Specification: Timing Output Control

**Feature Branch**: `002-timing-output-control`
**Created**: 2025-12-13
**Status**: Draft
**Input**: User description: "As a CLI user or developer, I want to control when timing output is displayed by enabling it via a `--timing` CLI flag (default false), so that I can avoid unnecessary timing information unless specifically requested and have consistent CLI output."

## Constitution Constraints _(mandatory)_

These are NON-NEGOTIABLE constraints derived from `.specify/memory/constitution.md`.

- **Interface-first**: Define interfaces before implementations; design for mocking.
- **Test-first**: Plan and write tests first; tests cover public APIs only (no white-box).
- **Package-first**: Business logic lives in `pkg/`; `cmd/` remains a thin wrapper.
- **Quality gates**: Work must pass `mockery`, `go test ./...`, `golangci-lint run`, `go build ./...`.
- **KISS/DRY/YAGNI**: Prefer simple, non-duplicated, non-speculative solutions.

## Clarifications

### Session 2025-12-13

- Q: When should timing output be emitted? → A: After each completed activity/stage.
- Q: Where should the `--timing` flag live? → A: Global/root persistent flag (available on all subcommands).
- Q: What format should durations use? → A: Go `time.Duration` string format.
- Q: What should happen on errors? → A: No timing output is printed.
- Q: How should the completion message + timing block be formatted? → A: Keep the existing `✔` completion message unchanged and print the timing block immediately after it, using the `⏲` glyph.

## User Scenarios & Testing _(mandatory)_

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.

  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Default Output (Priority: P1)

As a CLI user, I want timing output to be hidden by default, so normal CLI output stays clean and stable.

**Why this priority**: This is the default experience for all users and prevents output noise and regressions.

**Independent Test**: Run representative commands without enabling timing and assert that no timing lines are printed.

**Acceptance Scenarios**:

1. **Given** timing is not enabled via the CLI flag, **When** a user runs a command that completes successfully, **Then** no timing lines are displayed.
2. **Given** timing is not enabled, **When** a user runs a command that performs multiple internal activities, **Then** the output remains consistent with the same command output prior to this feature.

---

### User Story 2 - On-Demand Timing (Priority: P2)

As a CLI user or developer, I want to enable timing output for a single invocation using a `--timing` flag, so I can inspect performance without changing any files.

**Why this priority**: It enables quick, ad-hoc performance investigation.

**Independent Test**: Run a command with `--timing` enabled and assert the timing block is displayed in the expected format.

**Acceptance Scenarios**:

1. **Given** timing is enabled via the `--timing` flag, **When** an activity/stage completes successfully, **Then** the output includes a timing block matching:

   ```text
    ✔ completion message
    ⏲ current: <duration>
      total:  <duration>
   ```

   Where `current` is the time that the current activity took, and `total` is the accumulated time of all activities so far.

   And `<duration>` is formatted using Go `time.Duration` string formatting (e.g., `12ms`, `1.2s`, `3m4.5s`).

2. **Given** timing is enabled via the `--timing` flag, **When** multiple activities/stages complete within a single command run, **Then** each activity completion prints a timing block where `total` is monotonically non-decreasing across the run.

3. **Given** a user runs any command or subcommand with `--timing`, **When** the command performs one or more activities, **Then** timing output is enabled for that run.

4. **Given** timing is enabled, **When** an activity/stage completes successfully, **Then** the existing `✔` completion message remains unchanged and the timing block is printed immediately after it.

---

### Edge Cases

- Timing is enabled via `--timing` for a single run; the next run without `--timing` does not show timing.
- A command fails partway through execution: no timing output is printed (even if `--timing` is enabled).
- Commands with a single activity vs. multiple activities: `total` must always be greater than or equal to `current`.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: Timing output MUST be OFF by default.
- **FR-002**: Timing output MUST be enable-able via a global/root persistent `--timing` CLI flag (available on all commands and subcommands).
- **FR-003**: When timing is enabled, the CLI MUST display timing output immediately after each completed activity/stage’s existing `✔` completion message, and the timing block must include the `⏲` label:

  ```text
   ✔ completion message
   ⏲ current: <duration>
     total:  <duration>
  ```

- **FR-004**: `current` MUST represent the elapsed time for the most recently completed activity.
- **FR-005**: `total` MUST represent the accumulated elapsed time across all completed activities within the same command run.
- **FR-006**: `<duration>` MUST be formatted using Go `time.Duration` string formatting.
- **FR-007**: When timing is disabled, CLI output MUST NOT include any timing lines.
- **FR-008**: When a command run ends in an error, timing output MUST NOT be printed (even if `--timing` is enabled).
- **FR-009**: Documentation MUST describe the `--timing` flag and the default behavior (timing off).
- **FR-010**: Automated tests MUST cover: timing off by default, enabled via CLI flag, per-activity emission, error behavior, and output formatting.

### Assumptions

- A “command run” may consist of multiple “activities”; timing totals are scoped to a single run and reset between runs.
- The feature introduces an opt-in “enable timing” control via CLI flag only; a separate explicit “disable timing” CLI option is out of scope unless already present.

### Dependencies

- The CLI already tracks activity durations such that a “current” duration and an accumulated “total” can be derived.
- User-facing documentation exists and can be updated to describe new behavior.

### Out of Scope

- Changing what is timed (i.e., redefining activity boundaries) beyond what is necessary to report `current` and `total` consistently.
- Introducing a new dedicated “disable timing” CLI option unless it already exists.

### Key Entities _(include if feature involves data)_

- **Timing Preference**: Whether timing output is enabled for a command run (derived from the presence of the `--timing` flag).
- **Command Activity**: A user-visible unit of work that can be timed and reported as “current”.
- **Timing Summary**: The pair of durations reported to the user (`current`, `total`) for a command run.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: With default settings (no timing enabled), timing lines appear in 0% of command outputs.
- **SC-002**: With timing enabled, 100% of completed activities show both `current` and `total` in the specified format.
- **SC-003**: Users can enable timing via the `--timing` flag and observe the change in the same command run.
- **SC-004**: A user can follow documentation to enable timing in under 1 minute without additional guidance.
