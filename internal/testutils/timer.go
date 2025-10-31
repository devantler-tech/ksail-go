package testutils

import "time"

// RecordingTimer is a lightweight test implementation of the timer.Timer interface
// capturing the number of Start() and NewStage() calls and returning a fixed
// duration for deterministic snapshot output.
type RecordingTimer struct {
	StartCalls    int
	StartCount    int
	StageCalls    int
	NewStageCount int
}

func NewRecordingTimer() *RecordingTimer { return &RecordingTimer{} }

func (r *RecordingTimer) Start() {
	r.StartCalls++
	r.StartCount++
}

func (r *RecordingTimer) NewStage() {
	r.StageCalls++
	r.NewStageCount++
}
func (r *RecordingTimer) Stop() {}
func (r *RecordingTimer) GetTiming() (total time.Duration, stage time.Duration) {
	return time.Millisecond, time.Millisecond
}
