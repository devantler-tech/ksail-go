package errorhandler_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/errorhandler"
	"github.com/spf13/cobra"
)

func TestExecutorExecuteSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	executor := errorhandler.NewExecutor()

	err := executor.Execute(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecutorExecuteInvalidSubcommand(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "test"}
	root.AddCommand(&cobra.Command{Use: "valid"})
	root.SetArgs([]string{"invalid"})

	executor := errorhandler.NewExecutor()

	err := executor.Execute(root)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "unknown command \"invalid\" for \"test\"") {
		t.Fatalf("expected error message to contain unknown command text, got %q", message)
	}

	if strings.Contains(message, "Error: ") {
		t.Fatalf("expected message to strip 'Error:' prefix, got %q", message)
	}

	if !strings.Contains(message, "Run 'test --help' for usage.") {
		t.Fatalf("expected usage hint to be preserved, got %q", message)
	}
}

var errTestBoom = errors.New("boom")

func TestExecutorWithCustomNormalizer(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errTestBoom
		},
	}

	executor := errorhandler.NewExecutor(errorhandler.WithNormalizer(mockNormalizer{}))

	err := executor.Execute(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.HasPrefix(err.Error(), "custom boom") {
		t.Fatalf("expected custom normalized prefix, got %q", err.Error())
	}
}

type mockNormalizer struct{}

func (mockNormalizer) Normalize(string) string { return "custom boom" }
