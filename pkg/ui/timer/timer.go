// Package timer provides utilities for tracking command and stage execution times.
package timer

import "time"

// Timer tracks command start time and current stage timing.
type Timer struct {
	commandStart time.Time
	stageStart   time.Time
}

// New creates a new Timer and starts tracking from the current time.
func New() *Timer {
	now := time.Now()
	return &Timer{
		commandStart: now,
		stageStart:   now,
	}
}

// StartStage marks the start of a new stage.
// This resets the stage timer while preserving the command start time.
func (t *Timer) StartStage() {
	t.stageStart = time.Now()
}

// Total returns the duration since the timer was created.
func (t *Timer) Total() time.Duration {
	return time.Since(t.commandStart)
}

// Stage returns the duration since the last StartStage() call.
func (t *Timer) Stage() time.Duration {
	return time.Since(t.stageStart)
}
