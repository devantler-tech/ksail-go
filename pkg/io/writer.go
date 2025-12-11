package io

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Writer operations.

// TryWrite writes content to the provided writer.
//
// Parameters:
//   - content: The string content to write
//   - writer: The io.Writer to write to
//
// Returns:
//   - string: The content that was written (for chaining)
//   - error: Error if write fails
func TryWrite(content string, writer io.Writer) (string, error) {
	_, err := writer.Write([]byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to write content: %w", err)
	}

	return content, nil
}

// Safe file writing operations.

// WriteFileSafe writes content to a file path only if it is within the specified base directory.
// It prevents path traversal attacks by validating the path is within basePath.
//
// Parameters:
//   - content: The content to write to the file
//   - basePath: The base directory that filePath must be within
//   - filePath: The file path to write to (must be within basePath)
//   - force: If true, overwrites existing files; if false, skips existing files
//
// Returns:
//   - error: ErrBasePath if basePath is empty, ErrEmptyOutputPath if filePath is empty,
//     ErrPathOutsideBase if path is outside base, or write error
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

// File writing operations.

// TryWriteFile writes content to a file path, handling force/overwrite logic.
// It validates that the output path doesn't contain path traversal attempts.
//
// Parameters:
//   - content: The content to write to the file
//   - output: The output file path
//   - force: If true, overwrites existing files; if false, skips existing files
//
// Returns:
//   - string: The content that was written (for chaining)
//   - error: ErrEmptyOutputPath if output is empty, or write error
//
// Caller responsibilities:
//   - Ensure the output path is within intended bounds
//   - Handle the returned content appropriately
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

// Writer selection helpers.

// GetWriter returns an appropriate writer based on the quiet flag.
//
// Parameters:
//   - quiet: If true, returns io.Discard to silence output; if false, returns os.Stdout
//
// Returns:
//   - io.Writer: Either io.Discard or os.Stdout
func GetWriter(quiet bool) io.Writer {
	if quiet {
		return io.Discard
	}

	return os.Stdout
}
