package k3dprovisioner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	clustercommand "github.com/k3d-io/k3d/v5/cmd/cluster"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		return s.result, combineError(s.err, s.result)
	}
	return s.result, nil
}

func combineError(base error, res commandrunner.CommandResult) error {
	if base == nil {
		return nil
	}

	var details []string
	if trimmed := strings.TrimSpace(res.Stderr); trimmed != "" {
		details = append(details, trimmed)
	}
	if trimmed := strings.TrimSpace(res.Stdout); trimmed != "" {
		details = append(details, trimmed)
	}

	if len(details) == 0 {
		return base
	}

	return fmt.Errorf("%w: %s", base, strings.Join(details, " | "))
}

func TestCreateUsesConfigFlag(t *testing.T) {
	cfg := buildSimpleConfig("cfg-name")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "path/to/k3d.yaml", WithCommandRunner(runner))

	err := prov.Create(context.Background(), "")
	require.NoError(t, err)

	assert.ElementsMatch(
		t,
		[]string{"--config", "path/to/k3d.yaml", "cfg-name"},
		runner.recorded.args,
	)
}

func TestDeleteDefaultsToConfigName(t *testing.T) {
	cfg := buildSimpleConfig("from-config")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))

	err := prov.Delete(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, []string{"from-config"}, runner.recorded.args)
}

func TestStartUsesResolvedNameWithoutConfigFlag(t *testing.T) {
	cfg := buildSimpleConfig("cluster-a")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "path/to/config.yaml", WithCommandRunner(runner))

	err := prov.Start(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, []string{"cluster-a"}, runner.recorded.args)
}

func TestStopUsesExplicitName(t *testing.T) {
	cfg := buildSimpleConfig("cluster-a")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))

	err := prov.Stop(context.Background(), "custom")
	require.NoError(t, err)

	assert.Equal(t, []string{"custom"}, runner.recorded.args)
}

func TestListAddsJSONOutputFlag(t *testing.T) {
	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))

	_, err := prov.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"--output", "json"}, runner.recorded.args)
}

func TestListParsesJSON(t *testing.T) {
	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.result.Stdout = `[{"name":"alpha"},{"name":"beta"}]`

	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))
	names, err := prov.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta"}, names)
}

func TestListReturnsErrorWhenJSONInvalid(t *testing.T) {
	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.result.Stdout = `not-json`

	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))
	_, err := prov.List(context.Background())

	require.ErrorContains(t, err, "parse output")
}

func TestExistsReturnsFalseWhenNameEmpty(t *testing.T) {
	cfg := buildSimpleConfig("")
	runner := &stubRunner{}
	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))

	exists, err := prov.Exists(context.Background(), "")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCommandErrorsIncludeStdStreams(t *testing.T) {
	cfg := buildSimpleConfig("any")
	runner := &stubRunner{}
	runner.err = errors.New("boom")
	runner.result.Stdout = "stdout"
	runner.result.Stderr = "stderr"

	prov := NewK3dClusterProvisioner(cfg, "", WithCommandRunner(runner))
	err := prov.Create(context.Background(), "name")

	require.ErrorContains(t, err, "boom")
	require.ErrorContains(t, err, "stdout")
	require.ErrorContains(t, err, "stderr")
}

func TestCustomCommandBuilder(t *testing.T) {
	cfg := buildSimpleConfig("cfg")
	runner := &stubRunner{}
	builderCalls := 0

	prov := NewK3dClusterProvisioner(
		cfg,
		"",
		WithCommandRunner(runner),
		WithCommandBuilders(CommandBuilders{
			Create: func() *cobra.Command {
				builderCalls++
				return &cobra.Command{
					Run: func(cmd *cobra.Command, args []string) {
						fmt.Fprint(cmd.OutOrStdout(), "custom run")
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

func init() {
	// Ensure Cobra builder baseline is available to avoid nil pointers
	_ = clustercommand.NewCmdClusterCreate
}
