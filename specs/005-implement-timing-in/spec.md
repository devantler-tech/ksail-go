# Feature Specification: CLI Command Timing

**Feature Branch**: `005-implement-timing-in`
**Created**: 2025-10-01
**Status**: Implemented
**Input**: User description: "implement timing in the cli via a new package pkg/ui/timer. The timer should be used to estimate elapsed time of each command and its stages. A stage is defined by a new title. The timing must be printed in the following format [stage: x|total: x] where x is the time in go duration format. The timing must be printed in success messages. This is important to allow users to monitor how long commands and their individual stages take to run."

## Clarifications

### Session 2025-10-01

- Q: When should timing information be displayed to users? → A: After each stage completes (progressive updates throughout execution)
- Q: When a command fails mid-execution, should timing information be displayed in the error message? → A: No, only display timing on successful completion
- Q: At what duration threshold should the format switch between time units? → A: Always use Go's default Duration.String() (e.g., "1m30s", "500ms")
- Q: How should the timer mechanism integrate with the existing UI notification system? → A: Timer provides timing data; notify functions format and display it
- Q: Should timing be displayed for commands with only a single stage (no explicit title changes)? → A: Yes, but simplified as `[X]` (omit redundant stage time)

## User Scenarios & Testing *(mandatory)*

### Primary User Story

As a KSail user, when I run any CLI command (e.g., `ksail cluster up`, `ksail init`), I want to see how long the command takes and how long each stage within that command takes, so that I can monitor performance, identify slow operations, and understand where time is being spent during cluster provisioning and management operations.

### Acceptance Scenarios

1. **Given** a user runs `ksail cluster up` (multi-stage command), **When** the command completes successfully, **Then** the success message displays timing information in the format `[stage: 2m15s|total: 5m30s]` showing both the final stage duration and total elapsed time

2. **Given** a user runs `ksail init --distribution Kind` (single-stage command), **When** the command completes successfully, **Then** the success message displays timing in simplified format `[stage: 1.2s]` (note: implemented format includes the `stage:` label even for single-stage commands) with appropriate precision for sub-second operations

3. **Given** a command has multiple stages (e.g., "Initializing cluster", "Installing CNI", "Deploying workloads"), **When** each stage completes, **Then** timing information is displayed immediately showing `[stage: X|total: Y]` format, allowing users to monitor progress in real-time

4. **Given** a user runs any KSail command, **When** viewing the output, **Then** stage boundaries are clearly defined by title changes in the UI output

5. **Given** a long-running command with multiple stages, **When** intermediate stages complete, **Then** users can see the duration of each completed stage to understand progress

### Edge Cases

- What happens when a command completes in less than 1 millisecond? (timing should handle sub-millisecond precision)
- What happens when a stage has no explicit title change? (should default to tracking the command as a single stage and display timing in simplified `[stage: X]` format)
- What happens when a command fails mid-execution? (timing is tracked internally but NOT displayed to users; only successful completions show timing)
- How does timing behave with parallel operations within a command? (timing should track wall-clock time, not cumulative CPU time)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a timer utility that tracks elapsed time for CLI commands from start to completion

- **FR-002**: System MUST track elapsed time for individual stages within a command, where a stage is defined by a title change in the output

- **FR-003**: System MUST display timing information in format `[stage: X|total: Y]` for multi-stage commands, and simplified format `[stage: X]` for single-stage commands (where X is the last stage's elapsed time and Y is total elapsed time). (Note: Original draft proposed `[X]`; implementation standardizes on explicit `stage:` label in all cases.)

- **FR-004**: System MUST format timing durations using Go's standard Duration.String() method, which automatically produces appropriate units (e.g., "1m30s", "500ms", "2.5s")

- **FR-005**: System MUST allow commands to define stage boundaries by setting new titles during execution

- **FR-006**: System MUST maintain cumulative total time across all stages within a command execution

- **FR-007**: System MUST provide a mechanism to start timing when a command begins execution

- **FR-008**: System MUST provide a mechanism to mark stage transitions when titles change

- **FR-009**: System MUST provide a mechanism to stop timing when a command completes

- **FR-010**: System MUST format timing information to be human-readable and consistent across all commands

- **FR-011**: Timing information MUST be displayed after each stage completes during command execution, providing progressive updates to users

- **FR-012**: System MUST use Go's Duration.String() method which automatically handles sub-second durations with millisecond precision for operations <1s

- **FR-013**: System MUST handle long-running operations (minutes, hours) with appropriate formatting

- **FR-014**: System MUST NOT display timing information when commands fail or exit with errors; timing display is reserved for successful completions only

- **FR-015**: Timer MUST provide timing data as structured information (total duration, stage duration) that existing notification functions can format and display

- **FR-016**: System MUST use simplified timing format `[stage: X]` for single-stage commands to avoid redundant information (where total and stage time would be identical)

### Key Entities *(include if feature involves data)*

- **Timer**: Represents the timing tracker for a command execution. Tracks start time, current stage start time, and provides methods to calculate and return elapsed durations as structured data for the UI notification system to format and display.

- **Stage**: Represents a logical phase within a command execution, defined by a title. This is a conceptual entity tracked as state within the Timer (via `currentStage` field and `stageStartTime`), not a separate data structure.

- **Timing Metadata**: The formatted string containing stage and total timing information in the format `[stage: X|total: Y]`. Note: In technical implementation, the underlying data structure is called `TimingData` (Go struct), while this specification uses "Timing Metadata" as the user-facing concept.

## Review & Acceptance Checklist

> [!IMPORTANT]
> **GATE: Automated checks run during main() execution**

### Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Execution Status

> [!IMPORTANT]
> **Updated by main() during processing**

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (none identified)
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

## Notes

**User Value**: This feature directly addresses user feedback about understanding performance characteristics of KSail operations. Users running cluster provisioning operations (which can take several minutes) benefit from visibility into where time is spent, enabling them to:

1. Set appropriate expectations for command execution times
2. Identify performance bottlenecks or slow stages
3. Compare timing across different distributions (Kind vs K3d vs EKS)
4. Debug and troubleshoot slow operations

**Scope Boundaries**: This specification focuses solely on timing display for successful command completions. Error scenarios, performance optimization of the timing mechanism itself, and detailed profiling/tracing are out of scope for this feature.

**Testing Approach**: Acceptance tests should verify timing format correctness, stage boundary tracking, and consistent display across all commands. Performance tests should ensure the timing mechanism adds negligible overhead (<1ms) to command execution.

## Implementation Alignment

The implemented behavior matches all functional requirements with one intentional divergence from the original draft specification: the single-stage timing format uses `[stage: X]` instead of `[X]`. This change improves clarity and consistency across outputs. All multi-stage formats follow `[stage: X|total: Y]` as specified. Error paths correctly omit timing information. Duration formatting leverages Go's `time.Duration.String()` exactly.
