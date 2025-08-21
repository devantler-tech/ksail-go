package pathutils

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestExpandPath_ExpandsHomePrefix(t *testing.T) {
	t.Parallel()

	// Arrange
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}

	input := "~/some/nested/dir"

	// Act
	got, err := ExpandHomePath(input)
	if err != nil {
		t.Fatalf("ExpandHomePath returned error: %v", err)
	}

	want := filepath.Join(usr.HomeDir, "some", "nested", "dir")

	// Assert
	if got != want {
		t.Fatalf("ExpandHomePath(%q) = %q, want %q", input, got, want)
	}
}

func TestExpandPath_ReturnsUnchangedWhenNoTilde(t *testing.T) {
	t.Parallel()

	// Arrange
	cases := []string{
		filepath.Join("var", "tmp"),                              // relative path
		filepath.Join(string(filepath.Separator), "tmp", "file"), // absolute path
	}

	for _, inputPath := range cases {
		// Act
		got, err := ExpandHomePath(inputPath)
		if err != nil {
			t.Fatalf("ExpandHomePath returned error for %q: %v", inputPath, err)
		}

		// Assert
		if got != inputPath {
			t.Fatalf("ExpandHomePath(%q) = %q, want unchanged", inputPath, got)
		}
	}
}

func TestExpandPath_TildeOnlyUnchanged(t *testing.T) {
	t.Parallel()

	// Arrange
	input := "~" // No trailing slash, function should leave unchanged

	// Act
	got, err := ExpandHomePath(input)
	if err != nil {
		t.Fatalf("ExpandHomePath returned error: %v", err)
	}

	// Assert
	if got != input {
		t.Fatalf("ExpandHomePath(%q) = %q, want unchanged", input, got)
	}
}
