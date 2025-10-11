package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	errFactoryBoom       = errors.New("factory boom")
	errProvisionerFailed = errors.New("provisioner failed")
)

// setupLifecycleTest creates common test fixtures.
func setupLifecycleTest(
	t *testing.T,
) (*testutils.RecordingTimer, *testutils.StubFactory, *testutils.StubProvisioner, shared.LifecycleConfig) {
	t.Helper()

	timer := &testutils.RecordingTimer{}
	provisioner := &testutils.StubProvisioner{}
	factory := &testutils.StubFactory{
		Provisioner:        provisioner,
		DistributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	config := shared.LifecycleConfig{
		TitleEmoji:         "🚀",
		TitleContent:       "Test cluster...",
		ActivityContent:    "testing cluster",
		SuccessContent:     "cluster tested",
		ErrorMessagePrefix: "failed to test cluster",
		Action: func(ctx context.Context, prov clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return prov.Create(ctx, clusterName)
		},
	}

	return timer, factory, provisioner, config
}

// runLifecycleTest executes a lifecycle operation and returns the error.
func runLifecycleTest(
	t *testing.T,
	timer *testutils.RecordingTimer,
	factory *testutils.StubFactory,
	cfgManager *ksailconfigmanager.ConfigManager,
	config shared.LifecycleConfig,
) error {
	t.Helper()

	cmd, _ := testutils.NewCommand(t)
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}

	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		return fmt.Errorf("handle lifecycle: %w", err)
	}

	return nil
}

func TestLifecyclePattern_LoadConfigFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	timer, factory, _, config := setupLifecycleTest(t)

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(badPath, []byte(": invalid"), 0o600)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}

	err = shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err == nil || !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("expected load error, got %v", err)
	}

	if timer.StartCount != 1 || timer.NewStageCount != 0 {
		t.Fatalf(
			"expected StartCount=1, NewStageCount=0, got %d/%d",
			timer.StartCount,
			timer.NewStageCount,
		)
	}
}

func TestLifecyclePattern_FactoryFailure(t *testing.T) {
	t.Parallel()

	timer, _, _, config := setupLifecycleTest(t)
	factory := &testutils.StubFactory{Err: errFactoryBoom}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := runLifecycleTest(t, timer, factory, cfgManager, config)
	if err == nil || !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("expected factory error, got %v", err)
	}

	if timer.NewStageCount != 1 || factory.CallCount != 1 {
		t.Fatalf(
			"expected NewStageCount=1, CallCount=1, got %d/%d",
			timer.NewStageCount,
			factory.CallCount,
		)
	}
}

func TestLifecyclePattern_MissingProvisioner(t *testing.T) {
	t.Parallel()

	timer, _, _, config := setupLifecycleTest(t)
	factory := &testutils.StubFactory{}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := runLifecycleTest(t, timer, factory, cfgManager, config)
	if !errors.Is(err, shared.ErrMissingClusterProvisionerDependency) {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

func TestLifecyclePattern_ClusterNameFailure(t *testing.T) {
	t.Parallel()

	timer, _, _, config := setupLifecycleTest(t)
	factory := &testutils.StubFactory{
		Provisioner:        &testutils.StubProvisioner{},
		DistributionConfig: struct{}{},
	}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := runLifecycleTest(t, timer, factory, cfgManager, config)
	if err == nil || !strings.Contains(err.Error(), "failed to get cluster name") {
		t.Fatalf("expected cluster name error, got %v", err)
	}

	if factory.CallCount != 1 {
		t.Fatalf("expected CallCount=1, got %d", factory.CallCount)
	}
}

func TestLifecyclePattern_ProvisionerOperationFailure(t *testing.T) {
	t.Parallel()

	timer, factory, provisioner, config := setupLifecycleTest(t)
	provisioner.CreateErr = errProvisionerFailed
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := runLifecycleTest(t, timer, factory, cfgManager, config)
	if err == nil || !strings.Contains(err.Error(), "failed to test cluster") {
		t.Fatalf("expected operation error, got %v", err)
	}

	if provisioner.CreateCalls != 1 {
		t.Fatalf("expected CreateCalls=1, got %d", provisioner.CreateCalls)
	}
}

func TestLifecyclePattern_Success(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)
	timer, factory, provisioner, config := setupLifecycleTest(t)
	cfgManager := testutils.CreateConfigManager(t, out)

	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}

	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if timer.StartCount != 1 || timer.NewStageCount != 1 {
		t.Fatalf(
			"expected StartCount=1, NewStageCount=1, got %d/%d",
			timer.StartCount,
			timer.NewStageCount,
		)
	}

	if provisioner.CreateCalls != 1 {
		t.Fatalf("expected CreateCalls=1, got %d", provisioner.CreateCalls)
	}

	output := out.String()
	if !strings.Contains(output, "Test cluster...") || !strings.Contains(output, "cluster tested") {
		t.Fatalf("expected lifecycle messages in output, got %q", output)
	}
}
