package quiet_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
)

var (
	errMockOpen   = errors.New("mock open error")
	errNotDevNull = errors.New("mockFileOpener: only os.DevNull is allowed")
	errTest       = errors.New("test error")
)

// mockFileOpener is a test implementation of FileOpener that can simulate errors.
type mockFileOpener struct {
	shouldError bool
	errorMsg    string
}

func (m mockFileOpener) Open(name string) (*os.File, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%w: %s", errMockOpen, m.errorMsg)
	}
	// Only allow opening os.DevNull to avoid G304
	if name != os.DevNull {
		return nil, fmt.Errorf("%w, got %q", errNotDevNull, name)
	}

	f, err := os.Open(os.DevNull)
	if err != nil {
		return nil, fmt.Errorf("failed to open os.DevNull: %w", err)
	}

	return f, nil
}

func TestSilenceStdout_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	functionCalled := false
	testFunction := func() error {
		functionCalled = true

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
	expectedErr := errTest
	testFunction := func() error {
		return expectedErr
	}

	// Act
	err := quiet.SilenceStdout(testFunction)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSilenceStdoutWithOpener_OpenError(t *testing.T) {
	t.Parallel()

	// Arrange
	mockOpener := mockFileOpener{
		shouldError: true,
		errorMsg:    errMockOpen.Error(),
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

	if !errors.Is(err, errMockOpen) && err.Error() != "failed to open os.DevNull: mock open error" {
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
		err := file.Close()
		if err != nil {
			t.Errorf("failed to close file: %v", err)
		}
	}
}

func TestHandleCloseError_WithError(t *testing.T) {
	t.Parallel()

	// Arrange
	originalStderr := os.Stderr
	readPipe, writePipe, _ := os.Pipe()
	os.Stderr = writePipe

	testError := errTest

	// Act
	quiet.HandleCloseError(testError)

	// Clean up
	err := writePipe.Close()
	if err != nil {
		t.Errorf("failed to close writePipe: %v", err)
	}

	os.Stderr = originalStderr

	// Read what was written to stderr
	var buf bytes.Buffer

	_, err = buf.ReadFrom(readPipe)
	if err != nil {
		t.Errorf("failed to read from readPipe: %v", err)
	}

	err = readPipe.Close()
	if err != nil {
		t.Errorf("failed to close readPipe: %v", err)
	}

	// Assert
	expected := "failed to close os.DevNull: test error\n"
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
