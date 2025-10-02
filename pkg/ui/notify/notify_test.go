package notify_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	notify "github.com/devantler-tech/ksail-go/pkg/ui/notify"
)

func TestErrorf(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Errorf(&out, "%s: %d", "oops", 42)
	got := out.String()
	want := notify.ErrorSymbol + "oops: 42\n"

	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Error(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops"

	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestErrorln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Errorln(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops\n"

	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestWarnf(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Warnf(&out, "%s", "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarn(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Warn(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarnln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Warnln(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessf(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Successf(&out, "%s", "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccess(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Success(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Successln(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityf(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Activityf(&out, "%s", "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivity(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Activity(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Activityln(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestInfof(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Infof(&out, "%s", "details")
	got := out.String()
	want := notify.InfoSymbol + "details\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Info(&out, "details")
	got := out.String()
	want := notify.InfoSymbol + "details"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestInfoln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Infoln(&out, "details")
	got := out.String()
	want := notify.InfoSymbol + "details\n"

	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

// errorWriter is a mock writer that always returns an error.
type errorWriter struct{}

var errMockWrite = errors.New("mock write error")

func (ew errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("%w", errMockWrite)
}

func TestHandleNotifyErrorWithError(t *testing.T) {
	t.Parallel()

	// We'll capture stderr to verify the error handling
	oldStderr := os.Stderr
	readPipe, writePipe, _ := os.Pipe()
	os.Stderr = writePipe

	// Use an errorWriter to trigger the error path in handleNotifyError
	errWriter := errorWriter{}

	notify.Error(errWriter, "test message")

	// Restore stderr
	err := writePipe.Close()
	if err != nil {
		t.Fatalf("failed to close writePipe: %v", err)
	}

	os.Stderr = oldStderr

	// Read what was written to stderr
	buf := make([]byte, 1024)
	n, _ := readPipe.Read(buf)
	output := string(buf[:n])

	expectedErrorMsg := "notify: failed to print message: mock write error\n"
	if output != expectedErrorMsg {
		t.Fatalf("expected stderr output %q, got %q", expectedErrorMsg, output)
	}
}

func TestTitlef(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Titlef(&out, "🚀", "Starting %s version %s", "KSail", "v1.0.0")
	got := out.String()
	want := "🚀 Starting KSail version v1.0.0\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Title(&out, "🎯", "Deployment complete")
	got := out.String()
	want := "🎯 Deployment complete"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestTitleln(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	notify.Titleln(&out, "✨", "Process finished successfully")
	got := out.String()
	want := "✨ Process finished successfully\n"

	if got != want {
		t.Fatalf("output mismatch. want %q, got %q", want, got)
	}
}

func TestTitleFunctionsWithComplexEmojis(t *testing.T) {
	t.Parallel()

	t.Run("Titlef with complex emoji", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		notify.Titlef(&out, "🔧⚙️", "Configuration %s loaded", "production")
		got := out.String()
		want := "🔧⚙️ Configuration production loaded\n"

		if got != want {
			t.Fatalf("output mismatch. want %q, got %q", want, got)
		}
	})

	t.Run("Title with Unicode", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		notify.Title(&out, "📊", "Analytics dashboard ready")
		got := out.String()
		want := "📊 Analytics dashboard ready"

		if got != want {
			t.Fatalf("output mismatch. want %q, got %q", want, got)
		}
	})

	t.Run("Titleln with empty emoji", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		notify.Titleln(&out, "", "No emoji title")
		got := out.String()
		want := " No emoji title\n"

		if got != want {
			t.Fatalf("output mismatch. want %q, got %q", want, got)
		}
	})
}

// TestFormatTiming_IR002 validates timing format consistency (IR-002)
func TestFormatTiming_IR002(t *testing.T) {
	t.Run("Multi-stage format with different durations", func(t *testing.T) {
		total := 5*time.Minute + 30*time.Second
		stage := 2*time.Minute + 15*time.Second

		result := notify.FormatTiming(total, stage, true)
		expected := "[5m30s total|2m15s stage]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Single-stage format", func(t *testing.T) {
		duration := 1200 * time.Millisecond

		result := notify.FormatTiming(duration, duration, false)
		expected := "[1.2s]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Multi-stage with equal durations treated as single-stage", func(t *testing.T) {
		duration := 1 * time.Second

		result := notify.FormatTiming(duration, duration, true)
		expected := "[1s]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Sub-second precision", func(t *testing.T) {
		total := 500 * time.Millisecond
		stage := 200 * time.Millisecond

		result := notify.FormatTiming(total, stage, true)
		expected := "[500ms total|200ms stage]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Microsecond precision", func(t *testing.T) {
		duration := 123 * time.Microsecond

		result := notify.FormatTiming(duration, duration, false)
		expected := "[123µs]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Long duration format", func(t *testing.T) {
		total := 1*time.Hour + 23*time.Minute + 45*time.Second
		stage := 15 * time.Minute

		result := notify.FormatTiming(total, stage, true)
		expected := "[1h23m45s total|15m0s stage]"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

// TestSuccessWithTiming_IR003 validates optional timing display (IR-003)
func TestSuccessWithTiming_IR003(t *testing.T) {
	t.Run("Success without timing works as before", func(t *testing.T) {
		var out bytes.Buffer

		notify.Success(&out, "Cluster created")
		got := out.String()
		want := notify.SuccessSymbol + "Cluster created"

		if got != want {
			t.Errorf("Expected %q, got %q", want, got)
		}
	})

	t.Run("Success message with timing appended manually", func(t *testing.T) {
		var out bytes.Buffer

		// Pattern: Commands will manually append timing to message
		message := "Cluster created [5m30s total|2m15s stage]"
		notify.Success(&out, message)
		got := out.String()
		want := notify.SuccessSymbol + "Cluster created [5m30s total|2m15s stage]"

		if got != want {
			t.Errorf("Expected %q, got %q", want, got)
		}
	})

	t.Run("Successf with timing formatted in", func(t *testing.T) {
		var out bytes.Buffer

		// Pattern: Using Successf to append timing
		timing := "[2.5s]"
		notify.Successf(&out, "Cluster created %s", timing)
		got := out.String()
		want := notify.SuccessSymbol + "Cluster created [2.5s]\n"

		if got != want {
			t.Errorf("Expected %q, got %q", want, got)
		}
	})

	t.Run("Empty timing string handled gracefully", func(t *testing.T) {
		var out bytes.Buffer

		// If timing is empty, message should still work
		timing := ""
		message := fmt.Sprintf("Cluster created %s", timing)
		notify.Success(&out, message)
		got := out.String()
		want := notify.SuccessSymbol + "Cluster created "

		if got != want {
			t.Errorf("Expected %q, got %q", want, got)
		}
	})
}
