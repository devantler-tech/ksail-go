// Package io provides utilities for input and output operations.
package io

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// user read/write permission.
const filePermUserRW = 0600

// TryWrite writes content to the provided writer.
func TryWrite(content string, writer io.Writer) (string, error) {
	_, err := writer.Write([]byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to write content: %w", err)
	}

	return content, nil
}

// TryWriteFile writes content to a file path, handling force/overwrite logic.
// It uses the standard io.Writer interface and calls TryWrite internally.
func TryWriteFile(content string, output string, force bool) (string, error) {
	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(output)
		if err == nil {
			return content, nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to check file %s: %w", output, err)
		}
	}

	// Use os.OpenFile to get an io.Writer and call TryWrite
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermUserRW)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", output, err)
	}

	// Call TryWrite with the file writer
	result, writeErr := TryWrite(content, file)
	
	// Always close the file and capture any close error
	closeErr := file.Close()
	
	// Return write error if it occurred, otherwise return close error
	if writeErr != nil {
		return "", writeErr
	}

	if closeErr != nil {
		return "", fmt.Errorf("failed to close file %s: %w", output, closeErr)
	}
	
	return result, nil
}

// GetWriter returns an appropriate writer based on the quiet flag.
// If quiet is true, returns io.Discard to silence output.
// If quiet is false, returns os.Stdout for normal output.
func GetWriter(quiet bool) io.Writer {
	if quiet {
		return io.Discard
	}

	return os.Stdout
}
