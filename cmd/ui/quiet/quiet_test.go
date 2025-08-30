package quiet_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/ui/quiet"
	"github.com/stretchr/testify/assert"
)

func TestSilenceStdout_Success(t *testing.T) {
	// Test that output is indeed silenced
	var executed bool

	err := quiet.SilenceStdout(func() error {
		fmt.Println("This should not appear in test output")

		executed = true

		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed, "function should have been executed")
}

func TestSilenceStdout_Error(t *testing.T) {
	expectedErr := errors.New("test error")
	err := quiet.SilenceStdout(func() error {
		fmt.Println("This should not appear in test output")

		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestSilenceStdout_StdoutRestored(t *testing.T) {
	originalStdout := os.Stdout

	err := quiet.SilenceStdout(func() error {
		assert.NotEqual(t, originalStdout, os.Stdout, "stdout should be redirected during execution")

		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, originalStdout, os.Stdout, "stdout should be restored after execution")
}
