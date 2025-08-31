package quiet_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
)

func TestSilenceStdout_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	functionCalled := false
	testFunction := func() error {
		functionCalled = true
		// This would normally print to stdout, but should be silenced
		fmt.Println("This should not appear in the output")
		return nil
	}

	// Act
	err := quiet.SilenceStdout(testFunction)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !functionCalled {
		t.Error("expected test function to be called")
	}
	// Verify stdout was restored
	if os.Stdout == nil {
		t.Error("expected stdout to be restored")
	}
}

func TestSilenceStdout_FunctionError(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedErr := errors.New("test error")
	testFunction := func() error {
		return expectedErr
	}

	// Act
	err := quiet.SilenceStdout(testFunction)

	// Assert
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}