package quiet_test

import (
	"errors"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
)

var (
	errMockOpen = errors.New("mock open error")
	errTest     = errors.New("test error")
)

// setupMockFileOpener creates a mock file opener with pipe for testing.
func setupMockFileOpener(t *testing.T) (*quiet.MockFileOpener, *os.File) {
	t.Helper()
	mockOpener := quiet.NewMockFileOpener(t)
	
	// Create an in-memory pipe to avoid file system operations
	_, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for testing: %v", err)
	}
	
	mockOpener.EXPECT().Open(os.DevNull).Return(writeFile, nil)
	return mockOpener, writeFile
}

func TestSilenceStdout_Success(t *testing.T) {
	// Remove t.Parallel() to avoid race conditions with os.Stdout

	// Arrange
	mockOpener, writeFile := setupMockFileOpener(t)
	defer writeFile.Close()
	
	functionCalled := false
	testFunction := func() error {
		functionCalled = true
		return nil
	}

	// Act
	err := quiet.SilenceStdout(mockOpener, testFunction)

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
	// Remove t.Parallel() to avoid race conditions with os.Stdout

	// Arrange
	mockOpener, writeFile := setupMockFileOpener(t)
	defer writeFile.Close()
	
	expectedErr := errTest
	testFunction := func() error {
		return expectedErr
	}

	// Act
	err := quiet.SilenceStdout(mockOpener, testFunction)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSilenceStdout_OpenError(t *testing.T) {
	// Remove t.Parallel() to avoid race conditions with os.Stdout

	// Arrange
	mockOpener := quiet.NewMockFileOpener(t)
	mockOpener.EXPECT().Open(os.DevNull).Return(nil, errMockOpen)
	
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

	if !errors.Is(err, errMockOpen) {
		t.Errorf("expected wrapped mock open error, got %q", err.Error())
	}
}
