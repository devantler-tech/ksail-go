# Notify-Timer Integration Contract

## Contract Definition

This contract defines how the Timer package integrates with the existing `cmd/ui/notify` package to display timing information.

## Integration Pattern

**Pattern**: Separation of Concerns

- Timer provides raw timing data
- Notify formats and displays the data
- No direct dependency from timer → notify

## Notify Package Updates

### New Function: FormatTiming

```go
package notify

import (
    "fmt"
    "time"
)

// FormatTiming formats timing durations into display string
// Returns "[stage: X|total: Y]" for multi-stage or "[stage: X]" for single-stage
func FormatTiming(total, stage time.Duration, isMultiStage bool) string {
    if !isMultiStage || total == stage {
        return fmt.Sprintf("[stage: %s]", total.String())
    }
    return fmt.Sprintf("[stage: %s|total: %s]", stage.String(), total.String())
}
```

### Updated Function: Success

```go
package notify

// Success displays a success message with optional timing
// If timing string is provided, it's appended to the message
func Success(message string, timing ...string) {
    msg := message
    if len(timing) > 0 && timing[0] != "" {
        msg = fmt.Sprintf("%s %s", message, timing[0])
    }
    // ... existing success display logic
}
```

## Integration Requirements

### IR-001: Timer Independence

**Requirement**: Timer package MUST NOT import cmd/ui/notify package.

**Rationale**:

- Maintains clean architecture (pkg/ should not depend on cmd/)
- Allows timer to be used independently
- Follows constitutional dependency direction

**Verification**:

```bash
# Should not find any imports of cmd/ui/notify in pkg/ui/timer
grep -r "cmd/ui/notify" pkg/ui/timer/
# Expected: No matches
```

### IR-002: Timing Format Consistency

**Requirement**: FormatTiming() MUST produce formats matching specification.

**Expected Behavior**:

- Multi-stage: "[stage: 2m15s|total: 5m30s]"
- Single-stage: "[stage: 1.2s]"
- Uses Go's Duration.String() method

**Test Scenarios**:

```go
// Multi-stage
timing := FormatTiming(5*time.Minute + 30*time.Second, 2*time.Minute + 15*time.Second, true)
// Expect: "[stage: 2m15s|total: 5m30s]"

// Single-stage
timing := FormatTiming(1200*time.Millisecond, 1200*time.Millisecond, false)
// Expect: "[stage: 1.2s]"

// Multi-stage with equal durations (treated as single-stage)
timing := FormatTiming(1*time.Second, 1*time.Second, true)
// Expect: "[stage: 1s]"
```

### IR-003: Optional Timing Display

**Requirement**: Success() MUST handle cases with and without timing gracefully.

**Expected Behavior**:

- Without timing: `Success("Operation complete")` → "Operation complete"
- With timing: `Success("Operation complete", "[5s]")` → "Operation complete [5s]"
- Empty timing: `Success("Operation complete", "")` → "Operation complete"

**Test Scenarios**:

```go
// Test without timing
Success("Cluster created")
// Expect output: "Cluster created"

// Test with timing
Success("Cluster created", "[5m30s total|2m15s stage]")
// Expect output: "Cluster created [5m30s total|2m15s stage]"
```

### IR-004: Command Integration Pattern

**Requirement**: CLI commands MUST follow this integration pattern.

**Standard Pattern**:

```go
func RunCommand() error {
    // 1. Create and start timer
    timer := timer.New()
    timer.Start()

    // 2. Execute first stage
    notify.Title("Initializing")
    // ... do work ...

    // 3. Transition to new stages
    timer.NewStage("Deploying")
    notify.Title("Deploying")
    // ... do work ...

    // 4. On success, display with timing
    total, stage := timer.GetTiming()
    hasMultipleStages := true // or false for single-stage commands
    timingStr := notify.FormatTiming(total, stage, hasMultipleStages)
    notify.Success("Operation complete", timingStr)

    return nil
}
```

### IR-005: Error Cases - No Timing Display

**Requirement**: On command failure, timing MUST NOT be displayed.

**Expected Behavior**:

- Errors use notify.Error() without timing parameter
- Timer state is ignored on failure paths
- No timing cleanup needed (timer goes out of scope)

**Pattern**:

```go
func RunCommand() error {
    timer := timer.New()
    timer.Start()

    // Do work
    if err := doSomething(); err != nil {
        // NO timing display on error
        notify.Error("Operation failed: " + err.Error())
        return err
    }

    // Only on success
    total, stage := timer.GetTiming()
    timingStr := notify.FormatTiming(total, stage, true)
    notify.Success("Operation complete", timingStr)
    return nil
}
```

## Contract Test Implementation

Integration tests MUST be implemented to verify notify-timer integration.

### Test Location

- `pkg/ui/timer/integration_test.go` - Tests timer with real notify functions
- `cmd/ui/notify/notify_test.go` - Tests FormatTiming() function

### Test Structure

```go
func TestNotifyTimerIntegration(t *testing.T) {
    t.Run("IR-001: No circular dependency", func(t *testing.T) {
        // Verify imports using static analysis
    })

    t.Run("IR-002: Format consistency", func(t *testing.T) {
        // Test FormatTiming() with various durations
    })

    t.Run("IR-003: Optional timing", func(t *testing.T) {
        // Test Success() with and without timing
    })

    t.Run("IR-004: Command integration", func(t *testing.T) {
        // Test complete integration pattern
    })

    t.Run("IR-005: Error cases", func(t *testing.T) {
        // Verify no timing on errors
    })
}
```

## Success Criteria

Integration contracts MUST:

1. ✅ Maintain clean architecture (no pkg/ → cmd/ dependencies)
2. ✅ Produce correct timing formats
3. ✅ Handle optional timing display
4. ✅ Be demonstrated in working CLI commands
5. ✅ Pass all integration tests

## Contract Version

**Version**: 1.0.0
**Date**: 2025-10-01
**Status**: Draft (will be finalized during implementation)
