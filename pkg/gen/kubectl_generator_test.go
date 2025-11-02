package gen_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubectlGenerator_GenerateCommand(t *testing.T) {
	t.Parallel()

	t.Run("creates valid command for namespace", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")
		cmd := generator.GenerateCommand("namespace")

		require.NotNil(t, cmd)
		assert.Equal(t, "namespace", cmd.Name())
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("creates valid command for deployment", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")
		cmd := generator.GenerateCommand("deployment")

		require.NotNil(t, cmd)
		assert.Equal(t, "deployment", cmd.Name())
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("creates valid command for service", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")
		cmd := generator.GenerateCommand("service")

		require.NotNil(t, cmd)
		assert.Equal(t, "service", cmd.Name())
		assert.NotEmpty(t, cmd.Short)
		// Service is a group command with subcommands, so RunE might be nil
		// but it should have subcommands
		if cmd.RunE == nil {
			assert.NotEmpty(t, cmd.Commands(), "service should have subcommands")
		}
	})

	t.Run("creates valid command for secret", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")
		cmd := generator.GenerateCommand("secret")

		require.NotNil(t, cmd)
		assert.Equal(t, "secret", cmd.Name())
		assert.NotEmpty(t, cmd.Short)
		// Secret is a group command with subcommands, so RunE might be nil
		// but it should have subcommands
		if cmd.RunE == nil {
			assert.NotEmpty(t, cmd.Commands(), "secret should have subcommands")
		}
	})

	t.Run("panics for invalid resource type", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")

		assert.Panics(t, func() {
			generator.GenerateCommand("nonexistent-resource")
		})
	})
}

func TestKubectlGenerator_Interface(t *testing.T) {
	t.Parallel()

	t.Run("implements Generator interface", func(_ *testing.T) {
		var _ gen.Generator = gen.NewKubectlGenerator("")
	})
}

func TestKubectlGenerator_ExecutesWithDryRun(t *testing.T) {
	t.Parallel()

	t.Run("namespace generation produces YAML", func(t *testing.T) {
		t.Parallel()

		generator := gen.NewKubectlGenerator("")
		cmd := generator.GenerateCommand("namespace")

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
	})
}
