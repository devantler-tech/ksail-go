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

// setupMocks creates mock file opener and stdout manager with pipe for testing.
func setupMocks(t *testing.T) (*quiet.MockFileOpener, *quiet.MockStdoutManager, *os.File, *os.File) {
	t.Helper()
	mockOpener := quiet.NewMockFileOpener(t)
	mockStdoutManager := quiet.NewMockStdoutManager(t)
	
	// Create an in-memory pipe to avoid file system operations
	_, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for testing: %v", err)
	}
	
	// Create a fake stdout for testing
	_, fakeStdout, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create fake stdout for testing: %v", err)
	}
	
	mockOpener.EXPECT().Open(os.DevNull).Return(writeFile, nil)
	mockStdoutManager.EXPECT().GetStdout().Return(fakeStdout)
	mockStdoutManager.EXPECT().SetStdout(writeFile).Once()
	mockStdoutManager.EXPECT().SetStdout(fakeStdout).Once()
	
	return mockOpener, mockStdoutManager, writeFile, fakeStdout
}

func TestSilenceStdout_Success(t *testing.T) {
	t.Parallel() // Now safe to run in parallel with mocked stdout

	// Arrange
	mockOpener, mockStdoutManager, writeFile, fakeStdout := setupMocks(t)
	_ = writeFile // writeFile will be closed by SilenceStdout function

	defer func() {
		err := fakeStdout.Close()
		if err != nil {
			t.Errorf("failed to close fakeStdout: %v", err)
		}
	}()
	
	functionCalled := false
	testFunction := func() error {
		functionCalled = true

		return nil
	}

	// Act
	err := quiet.SilenceStdout(mockOpener, mockStdoutManager, testFunction)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !functionCalled {
		t.Error("expected test function to be called")
	}
}

func TestSilenceStdout_FunctionError(t *testing.T) {
	t.Parallel() // Now safe to run in parallel with mocked stdout

	// Arrange
	mockOpener, mockStdoutManager, writeFile, fakeStdout := setupMocks(t)
	_ = writeFile // writeFile will be closed by SilenceStdout function

	defer func() {
		err := fakeStdout.Close()
		if err != nil {
			t.Errorf("failed to close fakeStdout: %v", err)
		}
	}()
	
	expectedErr := errTest
	testFunction := func() error {
		return expectedErr
	}

	// Act
	err := quiet.SilenceStdout(mockOpener, mockStdoutManager, testFunction)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSilenceStdout_OpenError(t *testing.T) {
	t.Parallel() // Now safe to run in parallel with mocked stdout

	// Arrange
	mockOpener := quiet.NewMockFileOpener(t)
	mockStdoutManager := quiet.NewMockStdoutManager(t)

	mockOpener.EXPECT().Open(os.DevNull).Return(nil, errMockOpen)
	
	testFunction := func() error {
		return nil
	}

	// Act
	err := quiet.SilenceStdout(mockOpener, mockStdoutManager, testFunction)

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
