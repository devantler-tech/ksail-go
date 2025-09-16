package pathutils_test

import (
	"os/user"
	"path/filepath"
	"testing"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
)

func TestExpandHomePath(t *testing.T) {
	t.Parallel()

	usr, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands home prefix",
			input:    "~/some/nested/dir",
			expected: filepath.Join(usr.HomeDir, "some", "nested", "dir"),
		},
		{
			name:     "returns unchanged when no tilde - relative path",
			input:    filepath.Join("var", "tmp"),
			expected: filepath.Join("var", "tmp"),
		},
		{
			name:     "returns unchanged when no tilde - absolute path",
			input:    filepath.Join(string(filepath.Separator), "tmp", "file"),
			expected: filepath.Join(string(filepath.Separator), "tmp", "file"),
		},
		{
			name:     "tilde only unchanged",
			input:    "~",
			expected: "~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := pathutils.ExpandHomePath(tt.input)
			if err != nil {
				t.Fatalf("ExpandHomePath returned error: %v", err)
			}

			if got != tt.expected {
				t.Fatalf("ExpandHomePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
