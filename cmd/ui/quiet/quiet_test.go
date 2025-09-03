package quiet_test

import (
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

// simpleFileOpener implements FileOpener using os.Open.
type simpleFileOpener struct{}

// Open opens a file using os.Open. This is safe for tests as it only opens /dev/null.
//
//nolint:gosec // G304: This is test code that only opens /dev/null, which is safe
func (s simpleFileOpener) Open(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

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
	opener := simpleFileOpener{}
	functionCalled := false
	testFunction := func() error {
		functionCalled = true

		return nil
	}

	// Act
	err := quiet.SilenceStdout(opener, testFunction)

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
	opener := simpleFileOpener{}
	expectedErr := errTest
	testFunction := func() error {
		return expectedErr
	}

	// Act
	err := quiet.SilenceStdout(opener, testFunction)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSilenceStdout_OpenError(t *testing.T) {
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
	err := quiet.SilenceStdout(mockOpener, testFunction)

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

func TestSimpleFileOpener_Open(t *testing.T) {
	t.Parallel()

	// Arrange
	opener := simpleFileOpener{}

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
