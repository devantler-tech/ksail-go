package testutils

import "time"

// RecordingTimer is a lightweight test implementation of the timer.Timer interface
// that captures the number of Start() and NewStage() calls and returns a fixed
// duration for deterministic snapshot output.
type RecordingTimer struct {
	StartCalls    int
	StartCount    int
	StageCalls    int
	NewStageCount int
}

// Start records a Start invocation and increments counters.
func (r *RecordingTimer) Start() {
	r.StartCalls++
	r.StartCount++
}

// NewStage records a NewStage invocation and increments counters.
func (r *RecordingTimer) NewStage() {
	r.StageCalls++
	r.NewStageCount++
}

// Stop implements timer.Timer without additional behavior for tests.
func (r *RecordingTimer) Stop() {}

// GetTiming returns deterministic durations for snapshot-friendly assertions.
func (r *RecordingTimer) GetTiming() (time.Duration, time.Duration) {
	return time.Millisecond, time.Millisecond
}
