package timer

import "testing"

// TestStopCoverage ensures Stop() method is covered by tests.
// This internal test is needed because external tests (timer_test package)
// don't provide coverage for concrete implementation methods.
func TestStopCoverage(t *testing.T) {
	t.Parallel()

	tmr := New()
	tmr.Start()

	// Call Stop() to ensure it's covered
	tmr.Stop()

	// Verify timer still works after Stop()
	total, stage := tmr.GetTiming()
	if total == 0 || stage == 0 {
		t.Error("Expected non-zero durations after Stop()")
	}
}
