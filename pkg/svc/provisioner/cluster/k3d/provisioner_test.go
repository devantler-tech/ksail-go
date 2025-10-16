package k3dprovisioner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

type stubRunner struct {
	recorded struct {
		args []string
	}
	result commandrunner.CommandResult
	err    error
}

func (s *stubRunner) Run(
	_ context.Context,
	_ *cobra.Command,
	args []string,
) (commandrunner.CommandResult, error) {
	s.recorded.args = append([]string(nil), args...)
	if s.err != nil {
		mergeErr := commandrunner.MergeCommandError(s.err, s.result)

		return s.result, fmt.Errorf("merge command error: %w", mergeErr)
	}

	return s.result, nil
}

func TestCreateUsesConfigFlag(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("cfg-name")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"path/to/k3d.yaml",
		k3dprovisioner.WithCommandRunner(runner),
	)

	err := prov.Create(context.Background(), "")
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{"--config", "path/to/k3d.yaml", "cfg-name"},
		runner.recorded.args,
	)
}

func TestDeleteDefaultsToConfigName(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("from-config")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	err := prov.Delete(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, []string{"from-config"}, runner.recorded.args)
}

func TestStartUsesResolvedNameWithoutConfigFlag(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("cluster-a")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"path/to/config.yaml",
		k3dprovisioner.WithCommandRunner(runner),
	)

	err := prov.Start(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, []string{"cluster-a"}, runner.recorded.args)
}

func TestStopUsesExplicitName(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("cluster-a")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	err := prov.Stop(context.Background(), "custom")
	require.NoError(t, err)

	assert.Equal(t, []string{"custom"}, runner.recorded.args)
}

func TestListAddsJSONOutputFlag(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	_, err := prov.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"--output", "json"}, runner.recorded.args)
}

func TestListParsesJSON(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.result.Stdout = `[{"name":"alpha"},{"name":"beta"}]`

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)
	names, err := prov.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta"}, names)
}

func TestListReturnsErrorWhenJSONInvalid(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.result.Stdout = `not-json`

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)
	_, err := prov.List(context.Background())

	require.ErrorContains(t, err, "parse output")
}

func TestExistsReturnsFalseWhenNameEmpty(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("")
	runner := &stubRunner{}
	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	exists, err := prov.Exists(context.Background(), "")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCommandErrorsIncludeStdStreams(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.err = errBoom
	runner.result.Stdout = "stdout"
	runner.result.Stderr = "stderr"

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)
	err := prov.Create(context.Background(), "name")

	require.ErrorContains(t, err, "boom")
	require.ErrorContains(t, err, "stdout")
	require.ErrorContains(t, err, "stderr")
}

func TestCustomCommandBuilder(t *testing.T) {
	t.Parallel()

	cfg := buildSimpleConfig("cfg")
	runner := &stubRunner{}
	builderCalls := 0

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		cfg,
		"",
		k3dprovisioner.WithCommandRunner(runner),
		k3dprovisioner.WithCommandBuilders(k3dprovisioner.CommandBuilders{
			Create: func() *cobra.Command {
				builderCalls++
				testT := t

				return &cobra.Command{
					Run: func(cmd *cobra.Command, _ []string) {
						_, writeErr := fmt.Fprint(cmd.OutOrStdout(), "custom run")
						require.NoError(testT, writeErr)
					},
				}
			},
		}),
	)

	_, err := prov.List(context.Background())
	require.NoError(t, err)

	err = prov.Create(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, 1, builderCalls, "custom builder should be used once for create")
}

func buildSimpleConfig(name string) *v1alpha5.SimpleConfig {
	cfg := &v1alpha5.SimpleConfig{}
	cfg.Name = name

	return cfg
}
