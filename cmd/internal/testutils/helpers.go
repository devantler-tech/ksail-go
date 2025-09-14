// Package testutils provides testing helpers for command testing.
package testutils

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleCommandTestData holds test data for simple command testing.
type SimpleCommandTestData struct {
	CommandName   string
	NewCommand    func() *cobra.Command
	ExpectedUse   string
	ExpectedShort string
}

// TestSimpleCommandCreation tests command creation with common pattern.
func TestSimpleCommandCreation(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	cmd := data.NewCommand()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != data.ExpectedUse {
		t.Fatalf("expected Use to be %q, got %q", data.ExpectedUse, cmd.Use)
	}

	if cmd.Short != data.ExpectedShort {
		t.Fatalf("expected Short description to be %q, got %q", data.ExpectedShort, cmd.Short)
	}
}

// TestSimpleCommandExecution tests command execution with common pattern.
func TestSimpleCommandExecution(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// TestSimpleCommandHelp tests command help output with common pattern.
func TestSimpleCommandHelp(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// TestRunEError tests command RunE error handling with common pattern.
func TestRunEError(
	t *testing.T,
	runEFunc func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], args []string) error,
) {
	t.Helper()

	var out bytes.Buffer

	testCmd := &cobra.Command{
		Use:                    "",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "",
		GroupID:                "",
		Long:                   "",
		Example:                "",
		ValidArgs:              nil,
		ValidArgsFunction:      nil,
		Args:                   nil,
		ArgAliases:             nil,
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            nil,
		Version:                "",
		PersistentPreRun:       nil,
		PersistentPreRunE:      nil,
		PreRun:                 nil,
		PreRunE:                nil,
		Run:                    nil,
		RunE:                   nil,
		PostRun:                nil,
		PostRunE:               nil,
		PersistentPostRun:      nil,
		PersistentPostRunE:     nil,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: false,
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:         false,
			DisableNoDescFlag:         false,
			DisableDescriptions:       false,
			HiddenDefaultCmd:          false,
			DefaultShellCompDirective: nil,
		},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
	testCmd.SetOut(&out)

	mockManager := configmanager.NewMockConfigManager[v1alpha1.Cluster](t)
	mockManager.EXPECT().LoadConfig().Return(nil, ErrTestConfigLoadError)

	err := runEFunc(testCmd, mockManager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
}
