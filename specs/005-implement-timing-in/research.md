# Research: CLI Command Timing

## Overview

This document consolidates research findings for implementing timing functionality in KSail CLI commands.

## Decision: Timer Package Design Pattern

**Decision**: Implement timer as a stateful struct with Start(), NewStage(), and Stop() methods that track elapsed time.

**Rationale**:

- Aligns with Go idioms for managing stateful operations
- Simple API surface for CLI commands to integrate
- Encapsulates timing state (start time, stage time) internally
- Thread-safe through careful design (single-threaded CLI context)

**Alternatives Considered**:

1. **Functional approach** (passing time.Time values): Rejected because it exposes timing state management to callers, increasing complexity
2. **Channel-based approach**: Rejected as unnecessarily complex for synchronous CLI operations
3. **Global timer singleton**: Rejected as it violates clean architecture and testability principles

## Decision: Integration with Notify System

**Decision**: Timer provides `GetTiming() (total time.Duration, stage time.Duration)` method that notify functions call to format timing strings.

**Rationale**:

- Separation of concerns: timer tracks time, notify formats output
- Allows notify functions to control when/how timing is displayed
- Timer remains UI-agnostic (could be used in non-CLI contexts)
- Follows the constitutional principle of interface-based design

**Alternatives Considered**:

1. **Timer formats its own output**: Rejected because it couples timing logic with UI concerns
2. **Timer wraps notify functions**: Rejected because it creates inappropriate dependency direction (timer → notify)
3. **Middleware pattern**: Rejected as overengineering for this use case

## Decision: Stage Tracking Mechanism

**Decision**: CLI commands call `timer.NewStage(title string)` when changing operation phases. The timer records the previous stage's duration and resets the stage clock.

**Rationale**:

- Explicit stage boundaries make timing behavior predictable
- Commands control when stages change (when calling notify.Title())
- Stage titles provide context for timing display
- Simple API: one method call per stage transition

**Alternatives Considered**:

1. **Automatic stage detection**: Rejected because no reliable way to detect stage changes automatically
2. **Duration-based stages**: Rejected because stages are logical operations, not time-based
3. **Nested timer contexts**: Rejected as unnecessarily complex for linear command flows

## Decision: Timing Format

**Decision**: Use Go's `time.Duration.String()` method directly without custom formatting.

**Rationale**:

- Standard Go format (e.g., "1m30s", "500ms", "2.5s") is familiar to Go developers
- No need to maintain custom formatting logic
- Automatic precision handling (ms for <1s, s for <60s, m+s for longer)
- Consistent with other Go CLI tools

**Alternatives Considered**:

1. **Custom formatting** (always show 2 decimal places): Rejected as less readable for varied durations
2. **Humanize library** (e.g., "1 minute, 30 seconds"): Rejected as verbose for CLI output
3. **Unix timestamp format**: Rejected as not human-readable

## Decision: Error Handling Strategy

**Decision**: Timer continues tracking even on errors, but timing display only occurs on successful command completion.

**Rationale**:

- Timing data during failures adds noise to error messages
- Success messages are where users expect timing information
- Timer state is ephemeral (command-scoped), so no cleanup needed
- Simplifies error handling in commands

**Alternatives Considered**:

1. **Always display timing**: Rejected because it clutters error output
2. **Verbose flag for error timing**: Rejected as unnecessary complexity for v1
3. **Separate error timing display**: Rejected as out of scope per specification

## Best Practices: Go time Package

**Key Patterns**:

- Use `time.Now()` for wall-clock time (not `time.Since()` directly, to allow testing)
- Store `time.Time` values, calculate durations on-demand
- Avoid storing `time.Duration` values long-term (prefer start/end times)
- Use `time.Duration.String()` for human-readable output

**Testing Strategy**:

- Mock time via interface (allow injecting fake clock for tests)
- Test duration calculations independently from formatting
- Verify stage transitions don't lose cumulative time

## Best Practices: CLI Integration

**Key Patterns from KSail Architecture**:

- Commands use `cmd/ui/notify` package for all user-facing output
- Success messages use `notify.Success(message)` function
- Timing should be appended to success messages: `notify.Success(message, timing)`
- Stage changes occur when calling `notify.Title(newStage)`

**Integration Points**:

1. Command initialization: Create timer, call `Start()`
2. Stage transitions: Call `timer.NewStage(title)` when stage changes
3. Success completion: Get timing via `GetTiming()`, format, pass to `notify.Success()`

## Best Practices: Package-First Design (Constitutional)

**Requirements from Constitution**:

- Package must be in `pkg/` directory (not `cmd/` or `internal/`)
- Must be importable by external applications
- Must include comprehensive README.md
- Must include GoDoc comments on all exported symbols
- Must be independently testable (no CLI dependencies)

**Implementation**:

- Package location: `pkg/ui/timer`
- Public API: Timer interface + constructor
- No dependencies on `cmd/` packages
- Tests use standard `testing` package

## Research Summary

All technical unknowns have been resolved:

- ✅ Timer design pattern selected
- ✅ Integration mechanism with notify system defined
- ✅ Stage tracking approach specified
- ✅ Timing format decision made
- ✅ Error handling strategy determined
- ✅ Go time package best practices identified
- ✅ CLI integration patterns documented
- ✅ Constitutional requirements verified

**No remaining NEEDS CLARIFICATION items.**

Ready for Phase 1: Design & Contracts.
