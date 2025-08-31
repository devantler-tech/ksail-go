package quiet_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
)

// mockFileOpener is a test implementation of FileOpener that can simulate errors.
type mockFileOpener struct {
	shouldError bool
	errorMsg    string
}

func (m mockFileOpener) Open(name string) (*os.File, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return os.Open(name)
}

func TestSilenceStdout_Success(t *testing.T) {
	// Don't run in parallel since we're modifying os.Stdout

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
	// Don't run in parallel since we're modifying os.Stdout

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

func TestSilenceStdoutWithOpener_OpenError(t *testing.T) {
	t.Parallel()

	// Arrange
	mockOpener := mockFileOpener{
		shouldError: true,
		errorMsg:    "mock open error",
	}
	testFunction := func() error {
		return nil
	}

	// Act
	err := quiet.SilenceStdoutWithOpener(mockOpener, testFunction)

	// Assert
	if err == nil {
		t.Error("expected error when opener fails, got nil")
	}
	expectedSubstring := "failed to open os.DevNull"
	if len(err.Error()) < len(expectedSubstring) || err.Error()[:len(expectedSubstring)] != expectedSubstring {
		t.Errorf("expected error to start with %q, got %q", expectedSubstring, err.Error())
	}
	if !errors.Is(err, errors.New("mock open error")) && err.Error() != "failed to open os.DevNull: mock open error" {
		t.Errorf("expected wrapped error message, got %q", err.Error())
	}
}

func TestDefaultFileOpener_Open(t *testing.T) {
	t.Parallel()

	// Arrange
	opener := quiet.DefaultFileOpener{}

	// Act
	file, err := opener.Open(os.DevNull)

	// Assert
	if err != nil {
		t.Errorf("expected no error opening os.DevNull, got %v", err)
	}
	if file == nil {
		t.Error("expected file to be non-nil")
	}
	
	// Clean up
	if file != nil {
		file.Close()
	}
}

func TestHandleCloseError_WithError(t *testing.T) {
	// Don't run in parallel since we're modifying os.Stderr

	// Arrange
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	testError := errors.New("test close error")

	// Act
	quiet.HandleCloseError(testError)

	// Clean up
	w.Close()
	os.Stderr = originalStderr

	// Read what was written to stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	// Assert
	expected := "failed to close os.DevNull: test close error\n"
	if buf.String() != expected {
		t.Errorf("expected stderr output %q, got %q", expected, buf.String())
	}
}

func TestHandleCloseError_WithoutError(t *testing.T) {
	t.Parallel()

	// Arrange/Act
	quiet.HandleCloseError(nil)

	// Assert - This test doesn't modify global state, so nothing to assert
	// The fact that it doesn't panic or error is the test
}