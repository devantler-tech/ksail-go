# Timer Interface Contract

## Contract Definition

The Timer interface provides timing tracking functionality for CLI commands. This contract defines the expected behavior that any Timer implementation must satisfy.

## Interface Signature

```go
package timer

import "time"

// Timer tracks elapsed time for CLI command execution and stages
type Timer interface {
    // Start initializes the timer and begins tracking
    // Must be called before any other methods
    Start()

    // NewStage marks a transition to a new stage with the given title
    // Resets the stage timer while maintaining total elapsed time
    // Can be called multiple times during command execution
    NewStage(title string)

    // GetTiming returns the current total and stage elapsed durations
    // Can be called at any time after Start()
    // Returns (total duration, stage duration)
    GetTiming() (time.Duration, time.Duration)

    // Stop signals the end of timing (optional, for future extensibility)
    Stop()
}
```

## Contract Requirements

### CR-001: Start() Initialization

**Requirement**: Start() MUST initialize both total and stage timers to the current time.

**Expected Behavior**:

- After calling Start(), GetTiming() should return durations close to 0
- Calling Start() multiple times should reset the timer
- All other methods assume Start() has been called

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
total, stage := timer.GetTiming()
// Expect: total ≈ 0, stage ≈ 0
```

### CR-002: GetTiming() Before Start

**Requirement**: GetTiming() MUST return (0, 0) durations if called before Start(). It MUST NOT panic.

**Expected Behavior**:

- Always return (0, 0) durations if called before Start()
- Never panic

**Test Scenario**:

```go
timer := NewTimer()
total, stage := timer.GetTiming() // Before Start()
// Expect: (0, 0)
```

### CR-003: NewStage() Stage Transition

**Requirement**: NewStage() MUST reset the stage timer while preserving total elapsed time.

**Expected Behavior**:

- Total duration continues accumulating
- Stage duration resets to near-zero
- Multiple NewStage() calls work correctly

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
time.Sleep(100 * time.Millisecond)
timer.NewStage()
time.Sleep(50 * time.Millisecond)
total, stage := timer.GetTiming()
// Expect: total ≈ 150ms, stage ≈ 50ms
```

### CR-004: GetTiming() Returns Current State

**Requirement**: GetTiming() MUST return current elapsed durations without side effects.

**Expected Behavior**:

- Can be called multiple times
- Each call returns updated durations
- Does not modify timer state

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
time.Sleep(50 * time.Millisecond)
total1, stage1 := timer.GetTiming()
time.Sleep(50 * time.Millisecond)
total2, stage2 := timer.GetTiming()
// Expect: total2 > total1, stage2 > stage1
```

### CR-005: Single-Stage Command

**Requirement**: For commands without NewStage() calls, total and stage durations MUST be equal.

**Expected Behavior**:

- Start() → GetTiming() returns equal durations
- No NewStage() calls means single stage

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
time.Sleep(100 * time.Millisecond)
total, stage := timer.GetTiming()
// Expect: total ≈ stage ≈ 100ms
```

### CR-006: Duration Precision

**Requirement**: Durations MUST be precise enough to handle sub-millisecond operations.

**Expected Behavior**:

- time.Duration provides nanosecond precision
- Short operations (<1ms) should return non-zero durations

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
// Perform quick operation
total, stage := timer.GetTiming()
// Expect: total > 0, stage > 0 (even if < 1ms)
```

### CR-007: Stop() Idempotency

**Requirement**: Stop() SHOULD be idempotent and not affect GetTiming() results.

**Expected Behavior**:

- Stop() can be called safely
- GetTiming() after Stop() returns final durations
- Multiple Stop() calls are safe

**Test Scenario**:

```go
timer := NewTimer()
timer.Start()
time.Sleep(100 * time.Millisecond)
timer.Stop()
total1, stage1 := timer.GetTiming()
timer.Stop() // Second call
total2, stage2 := timer.GetTiming()
// Expect: total1 == total2, stage1 == stage2
```

## Contract Test Implementation

Contract tests MUST be implemented in `pkg/ui/timer/timer_test.go` and verify all requirements (CR-001 through CR-007).

### Test Structure

```go
package timer_test

import (
    "testing"
    "time"

    "github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func TestTimerContract(t *testing.T) {
    t.Run("CR-001: Start() initialization", func(t *testing.T) {
        // Test implementation
    })

    t.Run("CR-002: GetTiming() before Start()", func(t *testing.T) {
        // Test implementation
    })

    // ... additional contract tests
}
```

## Success Criteria

All contract tests MUST:

1. ✅ Pass when run via `go test ./pkg/ui/timer/...`
2. ✅ Be independent (no shared state)
3. ✅ Use testify assertions for clear failure messages
4. ✅ Include timing tolerances (±10ms) for sleep-based tests
5. ✅ Verify both success and edge cases

## Contract Version

**Version**: 1.0.0
**Date**: 2025-10-01
**Status**: Draft (will be finalized during implementation)
