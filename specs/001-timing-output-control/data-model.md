# Data Model: Timing Output Control

This feature introduces **no persisted data**. It adds a CLI-only toggle and a formatted output block.

## Entities

### 1) TimingSetting

- **Name**: `TimingSetting`
- **Type**: boolean
- **Source**: CLI flag (`--timing`)
- **Persistence**: none (invocation-scoped)
- **Default**: `false`

### 2) TimingSample

A single measurement emitted after a “timed activity” completes.

- **Name**: `TimingSample`
- **Fields**:
  - `current` (`time.Duration`): duration of the just-finished activity (maps to existing `timer.Timer` stage)
  - `total` (`time.Duration`): accumulated duration since invocation start (maps to existing `timer.Timer` total)

### 3) TimedActivity

- **Definition** (per spec): each progress/spinner step that emits a `✔ completion message`.
- **Start/End**:
  - “Start” is implicit (timer stage begins at `Start()` or `NewStage()`)
  - “End” is when the completion message is printed

## State / Transitions

- Invocation begins → `Timer.Start()`.
- After each activity completes → read `(total, stage)` via `Timer.GetTiming()` → emit timing output → `Timer.NewStage()` for the next activity.

## Validation Rules

- If the timer was never started, durations render as `0s` (existing `timer.Timer` behavior).
- Output formatting is stable and consistent across commands.
