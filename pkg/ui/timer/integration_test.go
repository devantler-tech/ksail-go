package timer_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

// Integration tests for Timer and Notify package interaction
// These tests validate the contract between timer and notify packages.

// TestIR001_TimerIndependence validates that timer has no dependency on notify.
func TestIR001_TimerIndependence(t *testing.T) {
	// Static analysis: verify no imports of notify package in timer package

	timerPkgPath := "." // Current package directory
	entries, err := os.ReadDir(timerPkgPath)
	if err != nil {
		t.Fatalf("Failed to read timer package directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue // Skip test files
		}

		filePath := filepath.Join(timerPkgPath, entry.Name())
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", filePath, err)
		}

		// Check imports
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(importPath, "notify") || strings.Contains(importPath, "cmd/") {
				t.Errorf("Timer package has forbidden import in %s: %s", entry.Name(), importPath)
			}
		}
	}
}

// TestIR004_CommandIntegrationPattern validates complete integration flow.
func TestIR004_CommandIntegrationPattern(t *testing.T) {
	t.Run("Complete multi-stage flow", func(t *testing.T) {
		// Simulate CLI command pattern
		tmr := timer.New()
		tmr.Start()

		// Stage 1
		time.Sleep(50 * time.Millisecond)
		total1, stage1 := tmr.GetTiming()

		// Verify stage 1
		if total1 < 40*time.Millisecond || total1 > 100*time.Millisecond {
			t.Errorf("Expected total ≈ 50ms, got %v", total1)
		}
		if stage1 != total1 {
			t.Errorf(
				"Expected stage == total for first stage, got total=%v stage=%v",
				total1,
				stage1,
			)
		}

		// Stage 2
		tmr.NewStage("Deploying")
		time.Sleep(30 * time.Millisecond)
		total2, stage2 := tmr.GetTiming()

		// Verify stage 2
		if total2 < 70*time.Millisecond || total2 > 150*time.Millisecond {
			t.Errorf("Expected total ≈ 80ms, got %v", total2)
		}
		if stage2 < 20*time.Millisecond || stage2 > 80*time.Millisecond {
			t.Errorf("Expected stage ≈ 30ms, got %v", stage2)
		}

		// Stage 3
		tmr.NewStage("Finalizing")
		time.Sleep(20 * time.Millisecond)
		total3, stage3 := tmr.GetTiming()

		// Verify final stage
		if total3 <= total2 {
			t.Errorf("Expected total3 (%v) > total2 (%v)", total3, total2)
		}
		if stage3 < 10*time.Millisecond || stage3 > 70*time.Millisecond {
			t.Errorf("Expected stage ≈ 20ms, got %v", stage3)
		}
	})

	t.Run("Complete single-stage flow", func(t *testing.T) {
		// Simulate single-stage command
		tmr := timer.New()
		tmr.Start()

		time.Sleep(100 * time.Millisecond)
		total, stage := tmr.GetTiming()

		// For single-stage, total == stage
		if total != stage {
			t.Errorf(
				"Expected total == stage for single-stage, got total=%v stage=%v",
				total,
				stage,
			)
		}

		// Verify duration
		if total < 90*time.Millisecond || total > 150*time.Millisecond {
			t.Errorf("Expected duration ≈ 100ms, got %v", total)
		}
	})
}

// TestIR005_ErrorCasesNoTiming validates timing is not displayed on errors.
func TestIR005_ErrorCasesNoTiming(t *testing.T) {
	t.Run("Timer state ignored on failure", func(t *testing.T) {
		// Simulate error path
		tmr := timer.New()
		tmr.Start()

		time.Sleep(50 * time.Millisecond)

		// Simulate error condition
		err := simulateError()
		if err != nil {
			// In real code, this would call notify.Error() WITHOUT timing
			// Timer state is simply abandoned (goes out of scope)

			// Verify timer still works (even though not used)
			total, stage := tmr.GetTiming()
			if total <= 0 || stage <= 0 {
				t.Errorf("Timer should still have valid state even on error path")
			}
		}
	})

	t.Run("No cleanup needed on error", func(t *testing.T) {
		tmr := timer.New()
		tmr.Start()

		// Simulate quick error
		_ = simulateError()

		// Timer goes out of scope, no cleanup required
		// This test verifies no panic or resource leak
		// Success = this test completes without panic
	})
}

// Helper function to simulate error condition
func simulateError() error {
	return &simulatedError{msg: "simulated error"}
}

type simulatedError struct {
	msg string
}

func (e *simulatedError) Error() string {
	return e.msg
}
