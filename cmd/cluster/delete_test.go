package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/shared"
	testutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	errFactoryFailure      = errors.New("factory failed")
	errDockerClientFailure = errors.New("docker client failure")
)

type (
	deleteConfigurator func(
		t *testing.T,
		cmd *cobra.Command,
		cfgManager *ksailconfigmanager.ConfigManager,
		configDir string,
	)
	deleteFactory func() *testutils.StubFactory
)

type deleteScenario struct {
	name                 string
	configure            deleteConfigurator
	factory              deleteFactory
	expectError          bool
	errMessage           string
	expectedFactoryCalls int
	expectedDeleteCalls  int
	expectCleanupWarning bool
}

func buildDeleteScenarios() []deleteScenario {
	return []deleteScenario{
		{
			name: "wraps_lifecycle_errors",
			factory: func() *testutils.StubFactory {
				return &testutils.StubFactory{Err: errFactoryFailure}
			},
			expectError:          true,
			errMessage:           "cluster deletion failed",
			expectedFactoryCalls: 1,
		},
		{
			name: "cleans_up_successfully",
			factory: func() *testutils.StubFactory {
				return &testutils.StubFactory{
					Provisioner:        &testutils.StubProvisioner{},
					DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
				}
			},
			expectedFactoryCalls: 1,
			expectedDeleteCalls:  1,
		},
		{
			name: "logs_cleanup_warning",
			configure: func(t *testing.T, _ *cobra.Command, _ *ksailconfigmanager.ConfigManager, configDir string) {
				t.Helper()
				ensureKindConfigHasPatch(t, filepath.Join(configDir, "kind.yaml"))
				stubDockerClientFailure(t, errDockerClientFailure)
			},
			factory: func() *testutils.StubFactory {
				return &testutils.StubFactory{
					Provisioner:        &testutils.StubProvisioner{},
					DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
				}
			},
			expectedFactoryCalls: 1,
			expectedDeleteCalls:  1,
			expectCleanupWarning: true,
		},
	}
}

func assertProvisionerDelete(t *testing.T, factory *testutils.StubFactory) {
	t.Helper()

	prov, ok := factory.Provisioner.(*testutils.StubProvisioner)
	require.True(t, ok, "expected stub provisioner")
	require.Equal(t, 1, prov.DeleteCalls)
}

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

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestHandleDeleteRunE(t *testing.T) {
	scenarios := buildDeleteScenarios()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			cmd, out, cfgManager, configDir := setupDeleteCommand(t)

			if scenario.configure != nil {
				scenario.configure(t, cmd, cfgManager, configDir)
			}

			factory := scenario.factory()

			err := handleDeleteRunE(cmd, cfgManager, sharedLifecycleDeps(factory))

			if scenario.expectError {
				require.Error(t, err)
				require.ErrorContains(t, err, scenario.errMessage)
			} else {
				require.NoError(t, err)
			}

			if scenario.expectedFactoryCalls > 0 {
				require.Equal(t, scenario.expectedFactoryCalls, factory.CallCount)
			}

			if scenario.expectedDeleteCalls > 0 {
				assertProvisionerDelete(t, factory)
			}

			if scenario.expectCleanupWarning {
				assert.Contains(t, out.String(), "failed to cleanup registries")
			} else if scenario.expectedDeleteCalls > 0 {
				assert.NotContains(t, out.String(), "failed to cleanup registries")
			}
		})
	}
}

func TestCleanupMirrorRegistries_IgnoresNonKindDistribution(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.Flags().Bool("delete-registry-volumes", false, "")

	cfg := v1alpha1.NewCluster()
	cfg.Spec.Distribution = v1alpha1.DistributionK3d

	err := cleanupMirrorRegistries(cmd, cfg, sharedLifecycleDeps(nil))

	require.NoError(t, err)
}

func TestCleanupMirrorRegistries_ReturnsKindConfigLoadError(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.Flags().Bool("delete-registry-volumes", false, "")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "kind.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(": invalid yaml"), 0o600))

	cfg := v1alpha1.NewCluster()
	cfg.Spec.Distribution = v1alpha1.DistributionKind
	cfg.Spec.DistributionConfig = configPath

	err := cleanupMirrorRegistries(cmd, cfg, sharedLifecycleDeps(nil))

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to load kind config")
}

func TestCleanupMirrorRegistries_ReturnsFlagLookupError(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	configDir := t.TempDir()

	kindPath := filepath.Join(configDir, "kind.yaml")
	writeKindWithPatch(t, kindPath)

	cfg := v1alpha1.NewCluster()
	cfg.Spec.Distribution = v1alpha1.DistributionKind
	cfg.Spec.DistributionConfig = kindPath

	err := cleanupMirrorRegistries(cmd, cfg, sharedLifecycleDeps(nil))

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get delete-registry-volumes flag")
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestCleanupMirrorRegistries_ReturnsDockerClientCreationError(t *testing.T) {
	cmd, _ := testutils.NewCommand(t)
	cmd.Flags().Bool("delete-registry-volumes", false, "")

	configDir := t.TempDir()
	kindPath := filepath.Join(configDir, "kind.yaml")
	writeKindWithPatch(t, kindPath)

	cfg := v1alpha1.NewCluster()
	cfg.Spec.Distribution = v1alpha1.DistributionKind
	cfg.Spec.DistributionConfig = kindPath

	stubDockerClientFailure(t, errDockerClientFailure)

	err := cleanupMirrorRegistries(cmd, cfg, sharedLifecycleDeps(nil))

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to create docker client")
}

func setupDeleteCommand(
	t *testing.T,
) (*cobra.Command, *bytes.Buffer, *ksailconfigmanager.ConfigManager, string) {
	t.Helper()

	cmd, out := testutils.NewCommand(t)
	cmd.SetContext(context.Background())
	cmd.Flags().Bool("delete-registry-volumes", false, "")

	cfgManager, configDir := newDeleteTestConfigManager(t, out)
	cfgManager.Viper.Set("spec.distributionConfig", filepath.Join(configDir, "kind.yaml"))

	return cmd, out, cfgManager, configDir
}

func newDeleteTestConfigManager(
	t *testing.T,
	writer *bytes.Buffer,
) (*ksailconfigmanager.ConfigManager, string) {
	t.Helper()

	tempDir := t.TempDir()
	testutils.WriteValidKsailConfig(t, tempDir)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	manager := ksailconfigmanager.NewConfigManager(writer, selectors...)
	manager.Viper.SetConfigFile(filepath.Join(tempDir, "ksail.yaml"))

	return manager, tempDir
}

func ensureKindConfigHasPatch(t *testing.T, path string) {
	t.Helper()

	const patch = "" +
		"containerdConfigPatches:\n" +
		"- |\n" +
		"  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]\n" +
		"    endpoint = [\"http://localhost:5000\"]\n"

	content, err := os.ReadFile(path) //nolint:gosec // test helper operates on generated file paths
	require.NoError(t, err, "failed to read kind config")

	if strings.Contains(string(content), "containerdConfigPatches") {
		return
	}

	content = append(content, []byte(patch)...)
	err = os.WriteFile(path, content, 0o600)
	require.NoError(t, err, "failed to update kind config")
}

func stubDockerClientFailure(t *testing.T, err error) {
	t.Helper()

	restore := shared.SetDockerClientFactoryForTest(func(...client.Opt) (*client.Client, error) {
		return nil, err
	})

	t.Cleanup(restore)
}

func writeKindWithPatch(t *testing.T, path string) {
	t.Helper()

	const content = "" +
		"kind: Cluster\n" +
		"apiVersion: kind.x-k8s.io/v1alpha4\n" +
		"name: kind\n" +
		"containerdConfigPatches:\n" +
		"- |\n" +
		"  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]\n" +
		"    endpoint = [\"http://localhost:5000\"]\n"

	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err, "failed to write kind config")
}

func sharedLifecycleDeps(
	factory *testutils.StubFactory,
) shared.LifecycleDeps {
	var tmr timer.Timer = &testutils.RecordingTimer{}

	var factoryInterface clusterprovisioner.Factory
	if factory != nil {
		factoryInterface = factory
	} else {
		factoryInterface = &testutils.StubFactory{}
	}

	return shared.LifecycleDeps{
		Timer:   tmr,
		Factory: factoryInterface,
	}
}
