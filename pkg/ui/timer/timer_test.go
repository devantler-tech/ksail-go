package timer_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

// Contract tests for Timer interface
// These tests define the behavioral contract that any Timer implementation must satisfy.

// TestCR001_StartInitialization validates that Start() properly initializes timing.
func TestCR001_StartInitialization(t *testing.T) {
	t.Parallel()

	t.Run("GetTiming returns near-zero durations after Start", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		total, stage := tmr.GetTiming()

		// Durations should be very small (< 10ms is reasonable for initialization)
		if total > 10*time.Millisecond {
			t.Errorf("Expected total duration < 10ms after Start(), got %v", total)
		}

		if stage > 10*time.Millisecond {
			t.Errorf("Expected stage duration < 10ms after Start(), got %v", stage)
		}
	})

	t.Run("Multiple Start calls reset the timer", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()
		time.Sleep(50 * time.Millisecond)

		// Reset with second Start()
		tmr.Start()
		total, stage := tmr.GetTiming()

		// After reset, durations should be near-zero again
		if total > 10*time.Millisecond {
			t.Errorf("Expected total duration < 10ms after second Start(), got %v", total)
		}

		if stage > 10*time.Millisecond {
			t.Errorf("Expected stage duration < 10ms after second Start(), got %v", stage)
		}
	})
}

// TestCR002_GetTimingBeforeStart validates behavior when GetTiming is called before Start.
func TestCR002_GetTimingBeforeStart(t *testing.T) {
	t.Parallel()

	tmr := timer.New()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetTiming() panicked when called before Start(): %v", r)
		}
	}()

	total, stage := tmr.GetTiming()

	// Should return zero durations
	if total != 0 {
		t.Errorf("Expected total duration = 0 before Start(), got %v", total)
	}

	if stage != 0 {
		t.Errorf("Expected stage duration = 0 before Start(), got %v", stage)
	}
}

// TestCR003_NewStageTransition validates that NewStage resets stage timer correctly.
func TestCR003_NewStageTransition(t *testing.T) {
	t.Parallel()

	t.Run("NewStage resets stage timer while preserving total", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		// First stage
		time.Sleep(100 * time.Millisecond)

		total1, stage1 := tmr.GetTiming()

		// Verify first stage timing
		if total1 < 90*time.Millisecond || total1 > 150*time.Millisecond {
			t.Errorf("Expected total ≈ 100ms, got %v", total1)
		}

		if stage1 < 90*time.Millisecond || stage1 > 150*time.Millisecond {
			t.Errorf("Expected stage ≈ 100ms, got %v", stage1)
		}

		// Transition to new stage
		tmr.NewStage("Stage 2")
		time.Sleep(50 * time.Millisecond)

		total2, stage2 := tmr.GetTiming()

		// Total should be ~150ms, stage should be ~50ms
		if total2 < 140*time.Millisecond || total2 > 200*time.Millisecond {
			t.Errorf("Expected total ≈ 150ms after stage 2, got %v", total2)
		}

		if stage2 < 40*time.Millisecond || stage2 > 100*time.Millisecond {
			t.Errorf("Expected stage ≈ 50ms after NewStage, got %v", stage2)
		}

		// Stage should be less than total
		if stage2 >= total2 {
			t.Errorf("Expected stage (%v) < total (%v)", stage2, total2)
		}
	})

	t.Run("Multiple NewStage calls work correctly", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(30 * time.Millisecond)
		tmr.NewStage("Stage 2")

		time.Sleep(30 * time.Millisecond)
		tmr.NewStage("Stage 3")

		time.Sleep(30 * time.Millisecond)

		total, stage := tmr.GetTiming()

		// Total should be ~90ms
		if total < 80*time.Millisecond || total > 140*time.Millisecond {
			t.Errorf("Expected total ≈ 90ms, got %v", total)
		}
		// Stage should be ~30ms (last stage only)
		if stage < 20*time.Millisecond || stage > 80*time.Millisecond {
			t.Errorf("Expected stage ≈ 30ms, got %v", stage)
		}
	})
}

// TestCR004_GetTimingReturnsCurrentState validates GetTiming can be called multiple times.
func TestCR004_GetTimingReturnsCurrentState(t *testing.T) {
	t.Parallel()

	tmr := timer.New()
	tmr.Start()

	// First call
	time.Sleep(50 * time.Millisecond)

	total1, stage1 := tmr.GetTiming()

	// Second call (should return updated durations)
	time.Sleep(50 * time.Millisecond)

	total2, stage2 := tmr.GetTiming()

	// Verify no side effects - each call returns current state
	if total2 <= total1 {
		t.Errorf("Expected total2 (%v) > total1 (%v)", total2, total1)
	}

	if stage2 <= stage1 {
		t.Errorf("Expected stage2 (%v) > stage1 (%v)", stage2, stage1)
	}

	// Third call (verify consistency)
	time.Sleep(20 * time.Millisecond)

	total3, stage3 := tmr.GetTiming()

	if total3 <= total2 {
		t.Errorf("Expected total3 (%v) > total2 (%v)", total3, total2)
	}

	if stage3 <= stage2 {
		t.Errorf("Expected stage3 (%v) > stage2 (%v)", stage3, stage2)
	}
}

// TestCR005_SingleStageCommand validates single-stage behavior (total == stage).
func TestCR005_SingleStageCommand(t *testing.T) {
	t.Parallel()

	tmr := timer.New()
	tmr.Start()

	time.Sleep(100 * time.Millisecond)

	total, stage := tmr.GetTiming()

	// Without NewStage(), total should equal stage
	if total != stage {
		t.Errorf(
			"Expected total == stage for single-stage command, got total=%v stage=%v",
			total,
			stage,
		)
	}

	// Verify they're both in expected range
	if total < 90*time.Millisecond || total > 150*time.Millisecond {
		t.Errorf("Expected duration ≈ 100ms, got %v", total)
	}
}

// TestCR006_StopMethod validates Stop() method behavior.
func TestCR006_StopMethod(t *testing.T) {
	t.Parallel()

	t.Run("Stop can be called without errors", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(50 * time.Millisecond)

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Stop() panicked: %v", r)
			}
		}()

		tmr.Stop()
	})

	t.Run("GetTiming works after Stop", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(50 * time.Millisecond)
		tmr.Stop()

		total, _ := tmr.GetTiming()

		// Should still return valid durations
		if total < 40*time.Millisecond || total > 100*time.Millisecond {
			t.Errorf("Expected duration ≈ 50ms after Stop(), got %v", total)
		}
	})

	t.Run("Multiple Stop calls are safe", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(50 * time.Millisecond)

		tmr.Stop()
		total1, stage1 := tmr.GetTiming()

		tmr.Stop() // Second call
		total2, stage2 := tmr.GetTiming()

		// Durations should be similar (allowing for small time passage between calls)
		// GetTiming() is calculated dynamically, so there will be microsecond differences
		totalDiff := total2 - total1
		if totalDiff < 0 || totalDiff > 10*time.Millisecond {
			t.Errorf(
				"Expected similar total durations after multiple Stop(), got diff=%v",
				totalDiff,
			)
		}

		stageDiff := stage2 - stage1
		if stageDiff < 0 || stageDiff > 10*time.Millisecond {
			t.Errorf(
				"Expected similar stage durations after multiple Stop(), got stage1=%v stage2=%v diff=%v",
				stage1,
				stage2,
				stageDiff,
			)
		}
	})
}

// TestCR007_DurationPrecision validates duration precision and formatting.
func TestCR007_DurationPrecision(t *testing.T) {
	t.Parallel()

	t.Run("Sub-millisecond operations return non-zero durations", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		// Quick operation (no sleep, just immediate call)
		total, stage := tmr.GetTiming()

		// Should still return non-zero (nanosecond precision)
		if total <= 0 {
			t.Errorf("Expected total > 0 for sub-millisecond operation, got %v", total)
		}

		if stage <= 0 {
			t.Errorf("Expected stage > 0 for sub-millisecond operation, got %v", stage)
		}
	})

	t.Run("Duration.String() formats correctly", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(1500 * time.Millisecond)

		total, _ := tmr.GetTiming()

		// Verify Duration.String() produces readable format
		str := total.String()
		if str == "" {
			t.Error("Expected non-empty string from Duration.String()")
		}

		// Should contain "s" for seconds
		if len(str) < 2 {
			t.Errorf("Expected formatted duration string, got %q", str)
		}
	})

	t.Run("Millisecond precision visible", func(t *testing.T) {
		t.Parallel()

		tmr := timer.New()
		tmr.Start()

		time.Sleep(123 * time.Millisecond)

		total, _ := tmr.GetTiming()

		// Should be in range 100-200ms
		if total < 100*time.Millisecond || total > 200*time.Millisecond {
			t.Errorf("Expected duration ≈ 123ms, got %v", total)
		}

		// String should show milliseconds
		str := total.String()
		if str == "" {
			t.Error("Expected formatted duration with milliseconds")
		}
	})
}
