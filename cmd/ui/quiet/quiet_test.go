package quiet_test

import (
	"io"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
)

func TestGetWriter_Quiet(t *testing.T) {
	t.Parallel()

	// Act
	writer := quiet.GetWriter(true)

	// Assert
	if writer != io.Discard {
		t.Errorf("expected io.Discard for quiet=true, got %T", writer)
	}
}

func TestGetWriter_NotQuiet(t *testing.T) {
	t.Parallel()

	// Act
	writer := quiet.GetWriter(false)

	// Assert
	if writer != os.Stdout {
		t.Errorf("expected os.Stdout for quiet=false, got %T", writer)
	}
}
