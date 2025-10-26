package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/stretchr/testify/assert"
)

func TestNewDeleteCmd(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewDeleteCmd(runtimeContainer)

	if cmd.Use != "delete" {
		t.Fatalf("expected Use to be 'delete', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short description to be set")
	}

	if cmd.RunE == nil {
		t.Fatal("expected RunE to be set")
	}

	var out bytes.Buffer
	cmd.SetOut(&out)
}

func TestNewDeleteLifecycleConfig(t *testing.T) {
	t.Parallel()

	config := newDeleteLifecycleConfig()

	assert.Equal(t, "🗑️", config.TitleEmoji)
	assert.Equal(t, "Delete cluster...", config.TitleContent)
	assert.Equal(t, "deleting cluster", config.ActivityContent)
	assert.Equal(t, "cluster deleted", config.SuccessContent)
	assert.Equal(t, "failed to delete cluster", config.ErrorMessagePrefix)
	assert.NotNil(t, config.Action)
}

func TestNewDeleteCmd_FlagConfiguration(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewDeleteCmd(runtimeContainer)

	flag := cmd.Flags().Lookup("delete-registry-volumes")
	assert.NotNil(t, flag, "delete-registry-volumes flag should be defined")
	assert.Equal(t, "false", flag.DefValue, "default value should be false")
}

func TestNewDeleteCommandRunE(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewDeleteCmd(runtimeContainer)

	assert.NotNil(t, cmd.RunE, "RunE handler should be set")
}

func TestDeleteLifecycleConfig_Action(t *testing.T) {
	t.Parallel()

	config := newDeleteLifecycleConfig()
	assert.NotNil(t, config.Action)

	// Test that the action function exists and has the correct signature
	// We can't execute it without a full setup, but we can verify it exists
}
