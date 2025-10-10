package shared_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	errFactoryError     = errors.New("factory error")
	errProvisionerError = errors.New("provisioner error")
)

// lifecycleTimer extends recordingTimer to track NewStage calls.
type lifecycleTimer struct {
	started       bool
	newStageCalls int
}

func (r *lifecycleTimer) Start()                                    { r.started = true }
func (r *lifecycleTimer) NewStage()                                 { r.newStageCalls++ }
func (r *lifecycleTimer) GetTiming() (time.Duration, time.Duration) { return 0, 0 }
func (r *lifecycleTimer) Stop()                                     {}

func TestHandleLifecycleRunE_ConfigLoadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(badPath, []byte(": invalid yaml"), 0o600)
	if err != nil {
		t.Fatalf("failed to write bad config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	timer := &lifecycleTimer{}
	factory := clusterprovisioner.NewMockFactory(t)
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}
	config := shared.LifecycleConfig{}

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	err = shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !timer.started {
		t.Error("expected timer to be started")
	}

	if timer.newStageCalls != 0 {
		t.Errorf("expected newStageCalls=0, got %d", timer.newStageCalls)
	}
}

func TestHandleLifecycleRunE_FactoryError(t *testing.T) {
	t.Parallel()

	cfgManager := createValidConfigManager(t)
	timer := &lifecycleTimer{}
	factory := clusterprovisioner.NewMockFactory(t)
	factory.On("Create", mock.Anything, mock.Anything).Return(nil, nil, errFactoryError).Once()
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}
	config := shared.LifecycleConfig{}

	err := runLifecycleHandlerTest(cfgManager, deps, config)
	if err == nil {
		t.Fatal("expected factory error")
	}

	if !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !timer.started || timer.newStageCalls != 1 {
		t.Errorf("expected started=true, newStageCalls=1, got started=%v, newStageCalls=%d",
			timer.started, timer.newStageCalls)
	}

	factory.AssertNumberOfCalls(t, "Create", 1)
}

func TestHandleLifecycleRunE_NilProvisioner(t *testing.T) {
	t.Parallel()

	cfgManager := createValidConfigManager(t)
	timer := &lifecycleTimer{}
	factory := clusterprovisioner.NewMockFactory(t)
	factory.On("Create", mock.Anything, mock.Anything).Return(nil, &v1alpha4.Cluster{}, nil).Once()
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}
	config := shared.LifecycleConfig{}

	err := runLifecycleHandlerTest(cfgManager, deps, config)
	if !errors.Is(err, shared.ErrMissingClusterProvisionerDependency) {
		t.Fatalf("expected ErrMissingClusterProvisionerDependency, got %v", err)
	}
}

func TestHandleLifecycleRunE_InvalidDistributionConfig(t *testing.T) {
	t.Parallel()

	cfgManager := createValidConfigManager(t)
	timer := &lifecycleTimer{}
	provisioner := clusterprovisioner.NewMockClusterProvisioner(t)
	factory := clusterprovisioner.NewMockFactory(t)
	factory.On("Create", mock.Anything, mock.Anything).Return(provisioner, struct{}{}, nil).Once()
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}
	config := shared.LifecycleConfig{}

	err := runLifecycleHandlerTest(cfgManager, deps, config)
	if err == nil {
		t.Fatal("expected error for invalid distribution config")
	}

	if !strings.Contains(err.Error(), "failed to get cluster name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleLifecycleRunE_ActionError(t *testing.T) {
	t.Parallel()

	cfgManager := createValidConfigManager(t)
	provisioner := clusterprovisioner.NewMockClusterProvisioner(t)
	provisioner.On("Create", mock.Anything, mock.Anything).Return(errProvisionerError).Once()
	deps := setupLifecycleDepsWithProvisioner(t, provisioner)
	config := createTestConfig()

	err := runLifecycleHandlerTest(cfgManager, deps, config)
	if err == nil {
		t.Fatal("expected action error")
	}

	if !strings.Contains(err.Error(), "test failed") {
		t.Fatalf("expected 'test failed' in error, got: %v", err)
	}

	provisioner.AssertNumberOfCalls(t, "Create", 1)
}

func TestHandleLifecycleRunE_Success(t *testing.T) {
	t.Parallel()

	cfgManager := createValidConfigManager(t)
	provisioner := clusterprovisioner.NewMockClusterProvisioner(t)
	provisioner.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	deps := setupLifecycleDepsWithProvisioner(t, provisioner)
	config := createTestConfig()

	timer, ok := deps.Timer.(*lifecycleTimer)
	if !ok {
		t.Fatal("expected timer to be *lifecycleTimer")
	}

	err := runLifecycleHandlerTest(cfgManager, deps, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !timer.started || timer.newStageCalls != 1 {
		t.Errorf("expected started=true, newStageCalls=1, got started=%v, newStageCalls=%d",
			timer.started, timer.newStageCalls)
	}

	provisioner.AssertNumberOfCalls(t, "Create", 1)
}

//nolint:paralleltest // Changes working directory
func TestNewLifecycleCommandWrapper_Success(t *testing.T) {
	dir := t.TempDir()
	writeValidConfig(t, filepath.Join(dir, "ksail.yaml"))

	t.Chdir(dir)

	var capturedTimer *lifecycleTimer

	var capturedFactory *clusterprovisioner.MockFactory

	runtimeContainer := createTestRuntimeWithDeps(t, &capturedTimer, &capturedFactory)

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	config := createTestLifecycleConfig()

	runE := shared.NewLifecycleCommandWrapper(runtimeContainer, cfgManager, config)

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := runE(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTimer == nil || !capturedTimer.started {
		t.Error("expected timer to be resolved and started")
	}

	if capturedFactory == nil {
		t.Fatal("expected factory to be resolved")
	}
	capturedFactory.AssertNumberOfCalls(t, "Create", 1)
}

func createTestRuntimeWithDeps(
	t *testing.T,
	capturedTimer **lifecycleTimer,
	capturedFactory **clusterprovisioner.MockFactory,
) *runtime.Runtime {
	return runtime.New(
		func(injector runtime.Injector) error {
			do.Provide(injector, func(do.Injector) (timer.Timer, error) {
				tmr := &lifecycleTimer{}
				*capturedTimer = tmr

				return tmr, nil
			})

			return nil
		},
		func(injector runtime.Injector) error {
			do.Provide(injector, func(do.Injector) (clusterprovisioner.Factory, error) {
				provisioner := clusterprovisioner.NewMockClusterProvisioner(t)
				provisioner.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
				
				factory := clusterprovisioner.NewMockFactory(t)
				factory.On("Create", mock.Anything, mock.Anything).
					Return(provisioner, &v1alpha4.Cluster{Name: "test"}, nil).Maybe()
				*capturedFactory = factory

				return factory, nil
			})

			return nil
		},
	)
}

func createTestLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🚀",
		TitleContent:       "Test...",
		ActivityContent:    "testing",
		SuccessContent:     "done",
		ErrorMessagePrefix: "failed",
		Action: func(ctx context.Context, prov clusterprovisioner.ClusterProvisioner, name string) error {
			return prov.Create(ctx, name)
		},
	}
}

func TestNewLifecycleCommandWrapper_TimerResolutionError(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.New()

	err := runWrapperTest(t, runtimeContainer)
	if err == nil {
		t.Fatal("expected timer resolution error")
	}

	if !strings.Contains(err.Error(), "resolve timer dependency") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewLifecycleCommandWrapper_FactoryResolutionError(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.New(func(injector runtime.Injector) error {
		do.Provide(injector, func(do.Injector) (timer.Timer, error) {
			return &lifecycleTimer{}, nil
		})

		return nil
	})

	err := runWrapperTest(t, runtimeContainer)
	if err == nil {
		t.Fatal("expected factory resolution error")
	}

	if !strings.Contains(err.Error(), "resolve provisioner factory dependency") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Helper functions

func runWrapperTest(t *testing.T, runtimeContainer *runtime.Runtime) error {
	t.Helper()

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	config := shared.LifecycleConfig{}

	runE := shared.NewLifecycleCommandWrapper(runtimeContainer, cfgManager, config)

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	return runE(cmd, nil)
}

func runLifecycleHandlerTest(
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
	config shared.LifecycleConfig,
) error {
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		return fmt.Errorf("lifecycle handler: %w", err)
	}

	return nil
}

func setupLifecycleDepsWithProvisioner(
	t *testing.T,
	provisioner clusterprovisioner.ClusterProvisioner,
) shared.LifecycleDeps {
	t.Helper()
	
	timer := &lifecycleTimer{}
	factory := clusterprovisioner.NewMockFactory(t)
	factory.On("Create", mock.Anything, mock.Anything).
		Return(provisioner, &v1alpha4.Cluster{Name: "test-cluster"}, nil).Maybe()

	return shared.LifecycleDeps{Timer: timer, Factory: factory}
}

func createTestConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🚀",
		TitleContent:       "Test operation...",
		ActivityContent:    "testing",
		SuccessContent:     "test complete",
		ErrorMessagePrefix: "test failed",
		Action: func(ctx context.Context, prov clusterprovisioner.ClusterProvisioner, name string) error {
			return prov.Create(ctx, name)
		},
	}
}

func createValidConfigManager(t *testing.T) *ksailconfigmanager.ConfigManager {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "ksail.yaml")
	writeValidConfig(t, configPath)

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(configPath)

	return cfgManager
}

func writeValidConfig(t *testing.T, path string) {
	t.Helper()

	const validConfig = `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
`

	err := os.WriteFile(path, []byte(validConfig), 0o600)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}
}
