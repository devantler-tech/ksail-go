//nolint:revive // Package name io is intentional for input/output operations
package io

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandHomePath expands a path beginning with ~/ to the user's home directory.
func ExpandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}

		return filepath.Join(usr.HomeDir, path[2:]), nil
	}

	return path, nil
}
