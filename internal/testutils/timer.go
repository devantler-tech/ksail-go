package testutils

import "time"

// RecordingTimer is a lightweight test implementation of the timer.Timer interface
// capturing the number of Start() and NewStage() calls and returning a fixed
// duration for deterministic snapshot output.
type RecordingTimer struct {
	StartCalls int
	StageCalls int
}

func (r *RecordingTimer) Start()    { r.StartCalls++ }
func (r *RecordingTimer) NewStage() { r.StageCalls++ }
func (r *RecordingTimer) Stop()     {}
func (r *RecordingTimer) GetTiming() (total time.Duration, stage time.Duration) {
	return time.Millisecond, time.Millisecond
}

// NewRecordingTimer constructs a new RecordingTimer.
func NewRecordingTimer() *RecordingTimer { return &RecordingTimer{} }
