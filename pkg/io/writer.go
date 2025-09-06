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

// WriteFileSafe writes content to a file path only if it is within the specified base directory.
// It prevents path traversal attacks by validating the resolved path is within baseDir.
func WriteFileSafe(content, basePath, filePath string, force bool) error {
	// Validate inputs
	if basePath == "" {
		return errors.New("base path cannot be empty")
	}
	if filePath == "" {
		return ErrEmptyOutputPath
	}

	// Clean and resolve paths to prevent path traversal
	basePath = filepath.Clean(basePath)
	filePath = filepath.Clean(filePath)

	// Ensure the file path is within the base directory
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(basePath, filePath)
	}

	// Get absolute paths to properly validate containment
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path %q: %w", basePath, err)
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path %q: %w", filePath, err)
	}

	// Check if the resolved file path is within the base directory
	relPath, err := filepath.Rel(absBasePath, absFilePath)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %q is outside %q", ErrPathOutsideBase, filePath, basePath)
	}

	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(absFilePath)
		if err == nil {
			return nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check file %s: %w", absFilePath, err)
		}
	}

	// Create directory if it doesn't exist (0750 permissions for security)
	dir := filepath.Dir(absFilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file safely with additional validation
	// The absFilePath has been validated against path traversal and is within allowed boundaries
	// #nosec G304 -- absFilePath is sanitized through comprehensive path validation above
	file, err := os.OpenFile(absFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermUserRW)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", absFilePath, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close file %s: %w", absFilePath, closeErr)
		}
	}()

	if _, err := file.Write([]byte(content)); err != nil {
		return fmt.Errorf("failed to write content to %s: %w", absFilePath, err)
	}

	return nil
}

// TryWriteFile writes content to a file path, handling force/overwrite logic.
// It uses the standard io.Writer interface and calls TryWrite internally.
// For enhanced security, it validates that the output path doesn't contain path traversal attempts.
func TryWriteFile(content string, output string, force bool) (string, error) {
	// Validate the output path cannot be empty
	if output == "" {
		return "", ErrEmptyOutputPath
	}

	// Clean the output path and check for path traversal attempts
	cleanOutput := filepath.Clean(output)
	
	// Check for obvious path traversal patterns
	if containsPathTraversal(cleanOutput) {
		return "", fmt.Errorf("%w: path contains traversal patterns", ErrPathOutsideBase)
	}

	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(cleanOutput)
		if err == nil {
			return content, nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to check file %s: %w", cleanOutput, err)
		}
	}

	// Create directory if it doesn't exist (0750 permissions for security)
	dir := filepath.Dir(cleanOutput)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Use os.OpenFile to get an io.Writer and call TryWrite
	file, err := os.OpenFile(cleanOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermUserRW)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", cleanOutput, err)
	}

	// Use defer with closure to ensure file is always closed and check for close errors
	var closeErr error

	defer func() {
		closeErr = file.Close()
	}()

	// Call TryWrite with the file writer
	result, writeErr := TryWrite(content, file)
	
	// Return write error if it occurred, otherwise return close error
	if writeErr != nil {
		return "", writeErr
	}

	if closeErr != nil {
		return "", fmt.Errorf("failed to close file %s: %w", cleanOutput, closeErr)
	}
	
	return result, nil
}

// containsPathTraversal checks if a path contains obvious traversal patterns.
func containsPathTraversal(path string) bool {
	// Check for ".." segments in the path
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}

	return false
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
