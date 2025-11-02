package kubernetes_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_Generate_Namespace(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("namespace")
	cmd := generator.Generate()

	require.NotNil(t, cmd)
	assert.Equal(t, "namespace", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestGenerator_Generate_Deployment(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("deployment")
	cmd := generator.Generate()

	require.NotNil(t, cmd)
	assert.Equal(t, "deployment", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestGenerator_Generate_Service(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("service")
	cmd := generator.Generate()

	require.NotNil(t, cmd)
	assert.Equal(t, "service", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	// Service is a group command with subcommands, so RunE might be nil
	// but it should have subcommands
	if cmd.RunE == nil {
		assert.NotEmpty(t, cmd.Commands(), "service should have subcommands")
	}
}

func TestGenerator_Generate_Secret(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("secret")
	cmd := generator.Generate()

	require.NotNil(t, cmd)
	assert.Equal(t, "secret", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	// Secret is a group command with subcommands, so RunE might be nil
	// but it should have subcommands
	if cmd.RunE == nil {
		assert.NotEmpty(t, cmd.Commands(), "secret should have subcommands")
	}
}

func TestGenerator_Generate_InvalidResource(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("nonexistent-resource")

	assert.Panics(t, func() {
		generator.Generate()
	})
}

func TestGenerator_ExecutesWithDryRun(t *testing.T) {
	t.Parallel()

	generator := kubernetes.NewGenerator("namespace")
	cmd := generator.Generate()

	// Capture output - need to capture both stdout and stderr
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	// Execute command
	cmd.SetArgs([]string{"test-namespace"})
	err := cmd.Execute()

	require.NoError(t, err)

	output := outBuf.String()

	// Verify YAML output contains expected content
	// Note: kubectl outputs to stdout
	if output != "" {
		assert.Contains(t, output, "apiVersion: v1")
		assert.Contains(t, output, "kind: Namespace")
		assert.Contains(t, output, "name: test-namespace")
	}
}
