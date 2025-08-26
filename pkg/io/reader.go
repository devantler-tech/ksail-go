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
