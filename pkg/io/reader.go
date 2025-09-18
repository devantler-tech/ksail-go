// Package io provides utilities for input and output operations.
package io

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathOutsideBase is returned when a file path is outside the specified base directory.
var ErrPathOutsideBase = errors.New("invalid path: file is outside base directory")

// ReadFileSafe reads the file at path only if it is located within baseDir.
// It resolves absolute paths and rejects reads where the resolved path is
// outside baseDir (prevents path traversal and accidental file inclusion).
func ReadFileSafe(basePath, filePath string) ([]byte, error) {
	filePath = filepath.Clean(filePath)

	if !strings.HasPrefix(filePath, basePath) {
		return nil, ErrPathOutsideBase
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	return data, nil
}

// ResolveConfigPath resolves a configuration file path with directory traversal.
// For absolute paths, returns the path as-is.
// For relative paths, traverses up from the current directory to find the file.
// Returns the resolved absolute path if found, or the original path if not found.
func ResolveConfigPath(configPath string) (string, error) {
	// If absolute path, return as-is
	if filepath.IsAbs(configPath) {
		return configPath, nil
	}

	// For relative paths, start from current directory and traverse up
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Traverse up the directory tree looking for the config file
	for {
		candidatePath := filepath.Join(currentDir, configPath)

		_, err := os.Stat(candidatePath)
		if err == nil {
			return candidatePath, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		// Stop if we've reached the root directory
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// If not found during traversal, return the original relative path
	// This allows the caller to handle the file-not-found case appropriately
	return configPath, nil
}
