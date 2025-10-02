package notify_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	notify "github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func TestWriteMessage_ErrorType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.ErrorType,
		Content: "test error",
		Writer:  &out,
	})

	got := out.String()
	want := "✗ test error\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_ErrorType_WithFormatting(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.ErrorType,
		Content: "error: %s (%d)",
		Args:    []any{"failed", 42},
		Writer:  &out,
	})

	got := out.String()
	want := "✗ error: failed (42)\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_WarningType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.WarningType,
		Content: "test warning",
		Writer:  &out,
	})

	got := out.String()
	want := "⚠ test warning\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_SuccessType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "test success",
		Writer:  &out,
	})

	got := out.String()
	want := "✔ test success\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_ActivityType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "test activity",
		Writer:  &out,
	})

	got := out.String()
	want := "► test activity\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_InfoType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.InfoType,
		Content: "test info",
		Writer:  &out,
	})

	got := out.String()
	want := "ℹ test info\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_TitleType(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "test title",
		Emoji:   "🚀",
		Writer:  &out,
	})

	got := out.String()
	want := "🚀 test title\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_TitleType_DefaultEmoji(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "test title with default emoji",
		Writer:  &out,
	})

	got := out.String()
	want := "ℹ️ test title with default emoji\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestWriteMessage_WithTimer(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	tmr := timer.New()
	tmr.Start()

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "operation complete",
		Timer:   tmr,
		Writer:  &out,
	})

	got := out.String()
	// Verify it has the success symbol and timing brackets
	if !strings.HasPrefix(got, "✔ operation complete [") {
		t.Fatalf("output should start with success symbol and have timing, got %q", got)
	}

	if !strings.Contains(got, "ms]") && !strings.Contains(got, "µs]") {
		t.Fatalf("output should contain timing in ms or µs, got %q", got)
	}
}

func TestWriteMessage_DefaultWriter(t *testing.T) {
	t.Parallel()

	// Test that nil writer defaults to stdout (just verify no panic)
	notify.WriteMessage(notify.Message{
		Type:    notify.InfoType,
		Content: "test with default writer",
		// Writer is nil - should default to os.Stdout
	})
	// If we get here without panicking, test passes
}

// TestFormatTiming_IR002 validates timing format consistency (IR-002).
func TestFormatTiming_IR002(t *testing.T) {
	t.Parallel()

	t.Run("Multi-stage format with different durations", func(t *testing.T) {
		t.Parallel()
		testMultiStageFormat(t)
	})

	t.Run("Single-stage format", func(t *testing.T) {
		t.Parallel()
		testSingleStageFormat(t)
	})

	t.Run("Multi-stage with equal durations treated as single-stage", func(t *testing.T) {
		t.Parallel()
		testEqualDurationsAsSingleStage(t)
	})

	t.Run("Sub-second precision", func(t *testing.T) {
		t.Parallel()
		testSubSecondPrecision(t)
	})

	t.Run("Microsecond precision", func(t *testing.T) {
		t.Parallel()
		testMicrosecondPrecision(t)
	})

	t.Run("Long duration format", func(t *testing.T) {
		t.Parallel()
		testLongDurationFormat(t)
	})
}

func testMultiStageFormat(t *testing.T) {
	t.Helper()

	total := 5*time.Minute + 30*time.Second
	stage := 2*time.Minute + 15*time.Second
	assertFormattedTiming(t, total, stage, true, "[5m30s total|2m15s stage]")
}

func testSingleStageFormat(t *testing.T) {
	t.Helper()

	duration := 1200 * time.Millisecond
	assertFormattedTiming(t, duration, duration, false, "[1.2s]")
}

func testEqualDurationsAsSingleStage(t *testing.T) {
	t.Helper()

	duration := 1 * time.Second
	assertFormattedTiming(t, duration, duration, true, "[1s]")
}

func testSubSecondPrecision(t *testing.T) {
	t.Helper()

	total := 500 * time.Millisecond
	stage := 200 * time.Millisecond
	assertFormattedTiming(t, total, stage, true, "[500ms total|200ms stage]")
}

func testMicrosecondPrecision(t *testing.T) {
	t.Helper()

	duration := 123 * time.Microsecond
	assertFormattedTiming(t, duration, duration, false, "[123µs]")
}

func testLongDurationFormat(t *testing.T) {
	t.Helper()

	total := 1*time.Hour + 23*time.Minute + 45*time.Second
	stage := 15 * time.Minute
	assertFormattedTiming(t, total, stage, true, "[1h23m45s total|15m0s stage]")
}

func assertFormattedTiming(
	t *testing.T,
	total, stage time.Duration,
	multiStage bool,
	expected string,
) {
	t.Helper()

	result := notify.FormatTiming(total, stage, multiStage)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
