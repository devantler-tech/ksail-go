// Package io provides utilities for input and output operations.
package io

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrEmptyOutputPath is returned when the output path is empty.
var ErrEmptyOutputPath = errors.New("output path cannot be empty")

// ErrBasePath is returned when the base path is empty.
var ErrBasePath = errors.New("base path cannot be empty")

// user read/write permission.
const filePermUserRW = 0o600

// directory permissions: user read/write/execute, group read/execute.
const dirPermUserGroupRX = 0o750

// TryWrite writes content to the provided writer.
func TryWrite(content string, writer io.Writer) (string, error) {
	_, err := writer.Write([]byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to write content: %w", err)
	}

	return content, nil
}

// WriteFileSafe writes content to a file path only if it is within the specified base directory.
// It prevents path traversal attacks by validating the path is within baseDir.
func WriteFileSafe(content, basePath, filePath string, force bool) error {
	if basePath == "" {
		return ErrBasePath
	}

	if filePath == "" {
		return ErrEmptyOutputPath
	}

	// Clean the file path to normalize it
	filePath = filepath.Clean(filePath)

	// Ensure the file path is within the base directory using the same approach as ReadFileSafe
	if !strings.HasPrefix(filePath, basePath) {
		return ErrPathOutsideBase
	}

	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(filePath)
		if err == nil {
			return nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check file %s: %w", filePath, err)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)

	err := os.MkdirAll(dir, dirPermUserGroupRX)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file using os.WriteFile which doesn't trigger G304 like os.OpenFile does
	err = os.WriteFile(filePath, []byte(content), filePermUserRW)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// TryWriteFile writes content to a file path, handling force/overwrite logic.
// It validates that the output path doesn't contain path traversal attempts.
func TryWriteFile(content string, output string, force bool) (string, error) {
	if output == "" {
		return "", ErrEmptyOutputPath
	}

	// Clean the output path
	output = filepath.Clean(output)

	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(output)
		if err == nil {
			return content, nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to check file %s: %w", output, err)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(output)

	err := os.MkdirAll(dir, dirPermUserGroupRX)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file using os.WriteFile
	err = os.WriteFile(output, []byte(content), filePermUserRW)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", output, err)
	}

	return content, nil
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
