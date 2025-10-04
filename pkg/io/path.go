package io

import (
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandHomePath expands a path beginning with ~/ to the user's home directory.
func ExpandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()

		return filepath.Join(usr.HomeDir, path[2:]), nil
	}

	return path, nil
}
