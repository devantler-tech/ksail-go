# Data Model: Timing Output Control

This feature does not introduce persisted data. It defines lightweight, in-memory concepts that shape CLI behavior.

## Entities

### Timing Preference

- Name: `TimingEnabled`
- Type: boolean
- Source: CLI flag `--timing`
- Scope: single command invocation
- Default: `false`

### Timing Summary

- Name: `TimingSummary`
- Fields:
  - `Current` (`time.Duration`): elapsed time for the most recently completed activity/stage
  - `Total` (`time.Duration`): accumulated elapsed time across completed activities in the current run
- Source of truth: `timer.Timer.GetTiming()`
- Validation rules:
  - `Total >= Current`
  - Both durations formatted using `time.Duration.String()` for display

### Command Activity

- Concept: a user-visible unit of work that produces an existing `✔` success/completion message.
- Relationship:
  - Each Activity completion may emit a `TimingSummary` block when `TimingEnabled == true`.

## State / Transitions

- Start of run:
  - `TimingEnabled` determined from CLI flag
  - `Total = 0`
- After each completed activity/stage:
  - `Current` updated to activity duration
  - `Total` monotonically non-decreasing
- End of run:
  - Timing state resets for the next invocation
