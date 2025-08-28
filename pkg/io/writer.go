// Package io provides utilities for input and output operations.
package io

import (
	"errors"
	"fmt"
	"os"
)

// user read/write permission.
const filePermUserRW = 0600

// FileWriter provides a reusable TryWrite helper for generators.
type FileWriter struct{}

// TryWrite writes content to opts.Output, handling force/overwrite messaging.
func (FileWriter) TryWrite(content string, output string, force bool) (string, error) {
	// Check if file exists and we're not forcing
	if !force {
		_, err := os.Stat(output)
		if err == nil {
			return content, nil // File exists and force is false, skip writing
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to check file %s: %w", output, err)
		}
	}

	err := os.WriteFile(output, []byte(content), filePermUserRW)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", output, err)
	}

	return content, nil
}
