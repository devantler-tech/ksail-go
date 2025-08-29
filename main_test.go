package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// Test that main builds and can show help
	if os.Getenv("BE_CRASHER") == "1" {
		main()
		return
	}

	// Test that the application builds correctly
	cmd := exec.Command(os.Args[0], "-test.run=TestMain")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	// The main function should exit with 0 when showing help successfully
	assert.NoError(t, err, "main() should complete without panic")
}

func TestMainHelp(t *testing.T) {
	// Test help command by running the binary
	if os.Getenv("BE_HELP_TESTER") == "1" {
		// Reset os.Args to simulate help flag
		os.Args = []string{"ksail", "--help"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelp")
	cmd.Env = append(os.Environ(), "BE_HELP_TESTER=1")
	output, err := cmd.CombinedOutput()
	
	// Help should exit with 0 and show usage
	assert.NoError(t, err, "help command should succeed")
	assert.Contains(t, string(output), "Usage:", "help output should contain usage info")
}

func TestVersion(t *testing.T) {
	// Test that version variables are properly set
	assert.NotEmpty(t, version, "version should be set")
	assert.NotEmpty(t, commit, "commit should be set")
	assert.NotEmpty(t, date, "date should be set")
}