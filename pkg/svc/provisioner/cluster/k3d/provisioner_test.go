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

//nolint:paralleltest
func TestCreateUsesConfigFlag(t *testing.T) {
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

//nolint:paralleltest
func TestDeleteDefaultsToConfigName(t *testing.T) {
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

//nolint:paralleltest
func TestStartUsesResolvedNameWithoutConfigFlag(t *testing.T) {
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

//nolint:paralleltest
func TestStopUsesExplicitName(t *testing.T) {
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

//nolint:paralleltest
func TestListAddsJSONOutputFlag(t *testing.T) {
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

//nolint:paralleltest
func TestListParsesJSON(t *testing.T) {
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

//nolint:paralleltest
func TestListReturnsErrorWhenJSONInvalid(t *testing.T) {
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

//nolint:paralleltest
func TestExistsReturnsFalseWhenNameEmpty(t *testing.T) {
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

//nolint:paralleltest
func TestCommandErrorsIncludeStdStreams(t *testing.T) {
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

//nolint:paralleltest
func TestCustomCommandBuilder(t *testing.T) {
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

//nolint:paralleltest // Updates shared runner stub state for sequential assertions.
func TestWithCommandBuildersOverridesAllCommands(t *testing.T) {
	recorder := &builderRecorder{}
	runner := &stubRunner{}
	runner.result.Stdout = `[{"name":"custom"}]`

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		buildSimpleConfig("cfg"),
		"",
		k3dprovisioner.WithCommandRunner(runner),
		k3dprovisioner.WithCommandBuilders(k3dprovisioner.CommandBuilders{
			Create: recorder.createBuilder,
			Delete: recorder.deleteBuilder,
			Start:  recorder.startBuilder,
			Stop:   recorder.stopBuilder,
			List:   recorder.listBuilder,
		}),
	)

	require.NoError(t, prov.Create(context.Background(), ""))
	require.NoError(t, prov.Delete(context.Background(), ""))
	require.NoError(t, prov.Start(context.Background(), ""))
	require.NoError(t, prov.Stop(context.Background(), ""))

	names, err := prov.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"custom"}, names)

	assert.Equal(t, 1, recorder.createCalls, "expected custom create builder to be used")
	assert.Equal(t, 1, recorder.deleteCalls, "expected custom delete builder to be used")
	assert.Equal(t, 1, recorder.startCalls, "expected custom start builder to be used")
	assert.Equal(t, 1, recorder.stopCalls, "expected custom stop builder to be used")
	assert.Equal(t, 1, recorder.listCalls, "expected custom list builder to be used")
}

type builderRecorder struct {
	createCalls int
	deleteCalls int
	startCalls  int
	stopCalls   int
	listCalls   int
}

func (b *builderRecorder) createBuilder() *cobra.Command {
	b.createCalls++

	return &cobra.Command{}
}

func (b *builderRecorder) deleteBuilder() *cobra.Command {
	b.deleteCalls++

	return &cobra.Command{}
}

func (b *builderRecorder) startBuilder() *cobra.Command {
	b.startCalls++

	return &cobra.Command{}
}

func (b *builderRecorder) stopBuilder() *cobra.Command {
	b.stopCalls++

	return &cobra.Command{}
}

func (b *builderRecorder) listBuilder() *cobra.Command {
	b.listCalls++

	return &cobra.Command{}
}

func TestExistsReturnsTrueForMatchingCluster(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{}
	runner.result.Stdout = `[{"name":"target"}]`

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		buildSimpleConfig("cfg"),
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	exists, err := prov.Exists(context.Background(), "target")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExistsPropagatesListErrors(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{}
	runner.err = errBoom

	prov := k3dprovisioner.NewK3dClusterProvisioner(
		buildSimpleConfig("cfg"),
		"",
		k3dprovisioner.WithCommandRunner(runner),
	)

	_, err := prov.Exists(context.Background(), "any")
	require.ErrorContains(t, err, "list")
}
