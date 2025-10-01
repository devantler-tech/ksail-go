# Data Model: CLI Command Timing

## Overview

This document defines the data structures and their relationships for the CLI command timing feature.

## Core Entities

### Timer

**Purpose**: Tracks elapsed time for a command execution and its stages.

**Attributes**:

- `startTime` (time.Time): The wall-clock time when command execution began
- `stageStartTime` (time.Time): The wall-clock time when the current stage began
- `currentStage` (string): The title of the current stage (empty for single-stage commands)

**Methods**:

- `Start()`: Initializes timing, sets startTime and stageStartTime to current time
- `NewStage(title string)`: Marks a stage transition, updates stageStartTime and currentStage
- `GetTiming() (total, stage time.Duration)`: Returns total elapsed time and current stage elapsed time
- `Stop()`: Optional cleanup method (for future extensibility)

**Validation Rules**:

- Start() must be called before any other methods
- NewStage() can be called multiple times (each resets stage timing)
- GetTiming() returns valid durations even before Stop() is called
- Stage titles should be non-empty strings (enforced by caller)

**State Transitions**:

```text
Created → Started (via Start())
Started → Started (via NewStage() - stage transition)
Started → Completed (via Stop() or end of command)
```

### TimingData

**Purpose**: Represents a snapshot of timing information at a point in time.

**Attributes**:

- `Total` (time.Duration): Total elapsed time since command start
- `Stage` (time.Duration): Elapsed time for current/last stage
- `StageTitle` (string): The title of the current/last stage (empty if single-stage)

**Validation Rules**:

- Total duration must be >= Stage duration
- Stage duration must be >= 0
- Both durations calculated from time.Time values, not stored directly

**Usage**:

- Returned by Timer.GetTiming()
- Consumed by notify functions for formatting
- Immutable snapshot (not updated after creation)

## Relationships

```text
Timer (1) ←→ (*) TimingData
  - Timer produces TimingData snapshots on demand
  - Each GetTiming() call creates a new TimingData
```

```text
Command (1) → (1) Timer
  - Each CLI command execution has one Timer instance
  - Timer lifecycle matches command lifecycle
```

```text
Timer (*) ← (1) Notify
  - Notify functions consume TimingData to format output
  - Timer has no knowledge of notify (dependency points inward)
```

## Type Definitions

### Timer Interface

```go
// Timer tracks elapsed time for CLI command execution
type Timer interface {
    // Start initializes timing tracking
    Start()

    // NewStage marks a stage transition with the given title
    NewStage(title string)

    // GetTiming returns total and stage elapsed durations
    GetTiming() (total, stage time.Duration)

    // Stop optionally signals completion (for future extensibility)
    Stop()
}
```

### TimingData Struct

```go
// TimingData represents timing information at a point in time
type TimingData struct {
    Total      time.Duration
    Stage      time.Duration
    StageTitle string
}
```

## Data Flow

1. **Command Start**: CLI command creates Timer, calls Start()
2. **Stage Transition**: Command calls NewStage(title) when operation phase changes
3. **Stage Completion**: Command calls GetTiming() to retrieve timing data
4. **Format Display**: Notify function formats timing into "[X total|Y stage]" or "[X]"
5. **Command End**: Timer goes out of scope (no persistence needed)

## Data Volume & Scale

- **Instances per execution**: 1 Timer per command invocation
- **Memory footprint**: ~100 bytes per Timer (3 time.Time + 1 string)
- **Lifetime**: Duration of single command execution (seconds to minutes)
- **Concurrency**: Not applicable (CLI commands execute sequentially)
- **Persistence**: None (timing is ephemeral, not stored)

## Testing Considerations

### Mockability

- Timer interface enables easy mocking via mockery
- Tests can inject fake time.Time values for deterministic testing
- GetTiming() returns calculated durations, not stored values

### Test Scenarios

1. **Single-stage timing**: Start() → GetTiming() → verify total == stage
2. **Multi-stage timing**: Start() → NewStage() → GetTiming() → verify total > stage
3. **Multiple stages**: Verify each NewStage() resets stage duration
4. **Duration precision**: Verify sub-millisecond operations handled correctly
5. **Zero duration**: Verify immediate GetTiming() returns ~0s durations

## Design Notes

**Why no Stage entity?**

- Stages are transient (not stored after transition)
- Only current stage matters for timing display
- Previous stage durations are not retained
- Keeping it simple aligns with constitutional simplicity principle

**Why time.Time instead of time.Duration?**

- Allows testing with injected clock
- More accurate (calculate on-demand vs accumulate errors)
- Standard Go pattern for timing operations

**Why no persistence layer?**

- Timing is for real-time user feedback only
- No requirement to analyze historical timing data
- Out of scope per specification (future enhancement possible)
