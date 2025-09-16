package pathutils_test

import (
	"os/user"
	"path/filepath"
	"testing"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
)

func TestExpandHomePathExpandsHomePrefix(t *testing.T) {
	t.Parallel()

	usr, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}

	input := "~/some/nested/dir"

	got, err := pathutils.ExpandHomePath(input)
	if err != nil {
		t.Fatalf("ExpandHomePath returned error: %v", err)
	}

	want := filepath.Join(usr.HomeDir, "some", "nested", "dir")

	if got != want {
		t.Fatalf("ExpandHomePath(%q) = %q, want %q", input, got, want)
	}
}

func TestExpandHomePathReturnsUnchangedWhenNoTilde(t *testing.T) {
	t.Parallel()

	cases := []string{
		filepath.Join("var", "tmp"),                              // relative path
		filepath.Join(string(filepath.Separator), "tmp", "file"), // absolute path
	}

	for _, inputPath := range cases {
		got, err := pathutils.ExpandHomePath(inputPath)
		if err != nil {
			t.Fatalf("ExpandHomePath returned error for %q: %v", inputPath, err)
		}

		if got != inputPath {
			t.Fatalf("ExpandHomePath(%q) = %q, want unchanged", inputPath, got)
		}
	}
}

func TestExpandHomePathTildeOnlyUnchanged(t *testing.T) {
	t.Parallel()

	input := "~" // No trailing slash, function should leave unchanged

	got, err := pathutils.ExpandHomePath(input)
	if err != nil {
		t.Fatalf("ExpandHomePath returned error: %v", err)
	}

	if got != input {
		t.Fatalf("ExpandHomePath(%q) = %q, want unchanged", input, got)
	}
}
