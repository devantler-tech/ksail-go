// Package pathutils provides utilities for manipulating filesystem paths.
package pathutils

import (
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandHomePath expands the home directory shortcut with the full path.
func ExpandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()

		return filepath.Join(usr.HomeDir, path[2:]), nil
	}

	return path, nil
}
