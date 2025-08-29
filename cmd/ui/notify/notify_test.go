package notify_test

import (
	"bytes"
	"testing"

	notify "github.com/devantler-tech/ksail-go/cmd/ui/notify"
)

// writer helpers no longer needed; we call *To(out, ...) variants directly

func TestErrorf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Errorf(&out, "%s: %d", "oops", 42)
	got := out.String()
	want := notify.ErrorSymbol + "oops: 42\n"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestError(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Error(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestErrorln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Errorln(&out, "oops")
	got := out.String()
	want := notify.ErrorSymbol + "oops\n"

	// Assert
	if got != want {
		t.Fatalf("stderr mismatch. want %q, got %q", want, got)
	}
}

func TestWarnf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warnf(&out, "%s", "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarn(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warn(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestWarnln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Warnln(&out, "careful")
	got := out.String()
	want := notify.WarningSymbol + "careful\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Successf(&out, "%s", "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccess(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Success(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestSuccessln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Successln(&out, "done")
	got := out.String()
	want := notify.SuccessSymbol + "done\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityf(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activityf(&out, "%s", "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivity(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activity(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}

func TestActivityln(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	// Act
	notify.Activityln(&out, "working")
	got := out.String()
	want := notify.ActivitySymbol + "working\n"

	// Assert
	if got != want {
		t.Fatalf("stdout mismatch. want %q, got %q", want, got)
	}
}
