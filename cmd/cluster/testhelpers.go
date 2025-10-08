package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errFactoryBoom = errors.New("factory boom")

// recordingTimer is a test timer that records calls
type recordingTimer struct {
	startCount    int
	newStageCount int
}

func (r *recordingTimer) Start()    { r.startCount++ }
func (r *recordingTimer) NewStage() { r.newStageCount++ }
func (r *recordingTimer) GetTiming() (time.Duration, time.Duration) {
	return 0, 0
}
func (r *recordingTimer) Stop() {}

// stubFactory is a test factory
type stubFactory struct {
	provisioner        clusterprovisioner.ClusterProvisioner
	distributionConfig any
	err                error
	callCount          int
}

//nolint:ireturn // Tests depend on returning the interface type.
func (s *stubFactory) Create(
	_ context.Context,
	_ *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	s.callCount++
	if s.err != nil {
		return nil, nil, s.err
	}

	return s.provisioner, s.distributionConfig, nil
}

func newCreateCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	return cmd, &out
}

func createConfigManager(t *testing.T, writer io.Writer) *ksailconfigmanager.ConfigManager {
	t.Helper()

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(writer, selectors...)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	cfgManager.Viper.SetConfigFile(filepath.Join(tempDir, "ksail.yaml"))

	return cfgManager
}

// testOperationCalls tracks calls to a specific provisioner operation.
type testOperationCalls interface {
	CallCount() int
}

// testHandlerBadConfigLoad tests handler behavior when config loading fails.
func testHandlerBadConfigLoad(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
) {
	t.Helper()
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	failingFactory := &stubFactory{err: nil}

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	const invalidYAML = ": invalid"
	const perm = 0o600

	err := os.WriteFile(badPath, []byte(invalidYAML), perm)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	deps := LifecycleDeps{Timer: timerStub, Factory: failingFactory}

	err = handler(cmd, cfgManager, deps)
	if err == nil {
		t.Fatal("expected configuration load error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("expected load error in message, got %q", err)
	}

	if timerStub.startCount != 1 {
		t.Fatalf("expected timer Start to be called once, got %d", timerStub.startCount)
	}

	if timerStub.newStageCount != 0 {
		t.Fatalf("expected timer NewStage to be skipped, got %d", timerStub.newStageCount)
	}
}

// testHandlerBadFactory tests handler behavior when factory fails.
func testHandlerBadFactory(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
) {
	t.Helper()
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{err: errFactoryBoom}
	cfgManager := createConfigManager(t, io.Discard)

	err := handler(cmd, cfgManager, LifecycleDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("expected factory failure message, got %q", err)
	}

	const expectedStageCount = 1
	if timerStub.newStageCount != expectedStageCount {
		t.Fatalf(
			"expected timer NewStage to be called before factory, got %d",
			timerStub.newStageCount,
		)
	}

	const expectedCallCount = 1
	if factory.callCount != expectedCallCount {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}
}

// testHandlerNilProvisioner tests handler behavior when provisioner is nil.
func testHandlerNilProvisioner(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
) {
	t.Helper()
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{}
	cfgManager := createConfigManager(t, io.Discard)

	err := handler(cmd, cfgManager, LifecycleDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected missing provisioner error, got nil")
	}

	if !strings.Contains(err.Error(), "missing cluster provisioner dependency") {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

// testHandlerBadClusterName tests handler behavior when cluster name retrieval fails.
func testHandlerBadClusterName(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
	provisionerStub clusterprovisioner.ClusterProvisioner,
) {
	t.Helper()
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{
		provisioner:        provisionerStub,
		distributionConfig: struct{}{},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := handler(cmd, cfgManager, LifecycleDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected cluster name error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get cluster name") {
		t.Fatalf("expected cluster name failure message, got %q", err)
	}

	const expectedCallCount = 1
	if factory.callCount != expectedCallCount {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}
}

// testHandlerOperationFails tests handler behavior when the operation fails.
func testHandlerOperationFails(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
	provisionerStub clusterprovisioner.ClusterProvisioner,
	expectedErr string,
	calls testOperationCalls,
) {
	t.Helper()
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{
		provisioner:        provisionerStub,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := handler(cmd, cfgManager, LifecycleDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected operation error, got nil")
	}

	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("expected %q in error, got %q", expectedErr, err)
	}

	const expectedCallCount = 1
	if calls.CallCount() != expectedCallCount {
		t.Fatalf("expected operation to be called once, got %d", calls.CallCount())
	}
}

// testHandlerSuccess tests successful handler execution.
func testHandlerSuccess(
	t *testing.T,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
	provisionerStub clusterprovisioner.ClusterProvisioner,
	expectedTitle string,
	expectedSuccess string,
	calls testOperationCalls,
) {
	t.Helper()
	t.Parallel()

	cmd, out := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{
		provisioner:        provisionerStub,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, out)

	err := handler(cmd, cfgManager, LifecycleDeps{Timer: timerStub, Factory: factory})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	const expected = 1
	if timerStub.startCount != expected || timerStub.newStageCount != expected {
		t.Fatalf(
			"expected timer Start/NewStage to be called once, got %d/%d",
			timerStub.startCount,
			timerStub.newStageCount,
		)
	}

	if calls.CallCount() != expected {
		t.Fatalf("expected operation to be called once, got %d", calls.CallCount())
	}

	output := out.String()
	if !strings.Contains(output, expectedTitle) {
		t.Fatalf("expected %q in output, got %q", expectedTitle, output)
	}

	if !strings.Contains(output, expectedSuccess) {
		t.Fatalf("expected %q in output, got %q", expectedSuccess, output)
	}
}

//nolint:funlen // Test helper complexity is acceptable
func testCmdIntegration(
	t *testing.T,
	newCmd func(*runtime.Runtime) *cobra.Command,
	provisionerStub clusterprovisioner.ClusterProvisioner,
	calls testOperationCalls,
) {
	t.Helper()

	var injectedTimer *recordingTimer

	factory := &stubFactory{
		provisioner:        provisionerStub,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}

	runtimeContainer := runtime.New(
		func(i runtime.Injector) error {
			do.Provide(i, func(runtime.Injector) (timer.Timer, error) {
				injectedTimer = &recordingTimer{}

				return injectedTimer, nil
			})

			return nil
		},
		func(i runtime.Injector) error {
			do.Provide(i, func(runtime.Injector) (clusterprovisioner.Factory, error) {
				return factory, nil
			})

			return nil
		},
	)

	cmd := newCmd(runtimeContainer)

	var out bytes.Buffer
	cmd.SetOut(&out)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)
	t.Chdir(tempDir)

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if injectedTimer == nil {
		t.Fatal("expected timer to be injected")
	}

	const expected = 1
	if injectedTimer.startCount == 0 {
		t.Fatalf("expected timer Start to be called, got %d", injectedTimer.startCount)
	}

	if factory.callCount != expected {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}

	if calls.CallCount() != expected {
		t.Fatalf("expected operation to be called once, got %d", calls.CallCount())
	}
}

// testCmdFactoryError tests command-level factory resolution error.
func testCmdFactoryError(
	t *testing.T,
	newCmd func(*runtime.Runtime) *cobra.Command,
) {
	t.Helper()

	// Runtime container with timer but no factory registered
	runtimeContainer := runtime.New(
		func(i runtime.Injector) error {
			do.Provide(i, func(runtime.Injector) (timer.Timer, error) {
				return &recordingTimer{}, nil
			})

			return nil
		},
	)

	cmd := newCmd(runtimeContainer)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)
	t.Chdir(tempDir)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected factory resolution error, got nil")
	}

	if !strings.Contains(err.Error(), "resolve provisioner factory dependency") {
		t.Fatalf("expected factory resolution error message, got %q", err)
	}
}


