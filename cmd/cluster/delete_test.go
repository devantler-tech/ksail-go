package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	errFactoryBoomDelete = errors.New("factory boom")
	errDeleteFailed      = errors.New("delete failed")
)

func TestHandleDeleteRunE_LoadConfigFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)

	timerStub := &testutils.RecordingTimer{}
	factoryCalled := 0
	failingFactory := &testutils.StubFactory{
		Err: nil,
	}

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(badPath, []byte(": invalid"), 0o600)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	deps := DeleteDeps{Timer: timerStub, Factory: failingFactory}

	err = HandleDeleteRunE(cmd, cfgManager, deps)
	if err == nil {
		t.Fatal("expected configuration load error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("expected load error in message, got %q", err)
	}

	if timerStub.StartCount != 1 {
		t.Fatalf("expected timer Start to be called once, got %d", timerStub.StartCount)
	}

	if timerStub.NewStageCount != 0 {
		t.Fatalf("expected timer NewStage to be skipped, got %d", timerStub.NewStageCount)
	}

	if factoryCalled != 0 {
		t.Fatalf("expected factory not to be invoked, got %d", factoryCalled)
	}
}

func TestHandleDeleteRunE_FactoryFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	factory := &testutils.StubFactory{Err: errFactoryBoomDelete}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("expected factory failure message, got %q", err)
	}

	if timerStub.NewStageCount != 1 {
		t.Fatalf(
			"expected timer NewStage to be called before factory, got %d",
			timerStub.NewStageCount,
		)
	}

	if factory.CallCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.CallCount)
	}
}

func TestHandleDeleteRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	factory := &testutils.StubFactory{}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if !errors.Is(err, errMissingClusterProvisionerForDelete) {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

func TestHandleDeleteRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	factory := &testutils.StubFactory{
		Provisioner:        &testutils.StubProvisioner{},
		DistributionConfig: struct{}{},
	}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected cluster name error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get cluster name") {
		t.Fatalf("expected cluster name failure message, got %q", err)
	}

	if factory.CallCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.CallCount)
	}
}

func TestHandleDeleteRunE_ReturnsErrorWhenProvisionerDeleteFails(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	provisioner := &testutils.StubProvisioner{DeleteErr: errDeleteFailed}
	factory := &testutils.StubFactory{
		Provisioner:        provisioner,
		DistributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := testutils.CreateConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected provisioner delete error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete cluster") {
		t.Fatalf("expected delete failure message, got %q", err)
	}

	if provisioner.DeleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.DeleteCalls)
	}
}

func TestHandleDeleteRunE_Success(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	provisioner := &testutils.StubProvisioner{}
	factory := &testutils.StubFactory{
		Provisioner:        provisioner,
		DistributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := testutils.CreateConfigManager(t, out)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if timerStub.StartCount != 1 || timerStub.NewStageCount != 1 {
		t.Fatalf(
			"expected timer Start/NewStage to be called once, got %d/%d",
			timerStub.StartCount,
			timerStub.NewStageCount,
		)
	}

	if provisioner.DeleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.DeleteCalls)
	}

	output := out.String()
	if !strings.Contains(output, "Delete cluster...") {
		t.Fatalf("expected deletion title in output, got %q", output)
	}

	if !strings.Contains(output, "cluster deleted") {
		t.Fatalf("expected success message in output, got %q", output)
	}
}

//nolint:paralleltest
func TestNewDeleteCmd_RunESuccess(t *testing.T) {
	var injectedTimer *testutils.RecordingTimer

	provisioner := &testutils.StubProvisioner{}
	factory := &testutils.StubFactory{
		Provisioner:        provisioner,
		DistributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}

	runtimeContainer := runtime.New(
		func(i runtime.Injector) error {
			do.Provide(i, func(runtime.Injector) (timer.Timer, error) {
				injectedTimer = &testutils.RecordingTimer{}

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

	cmd := NewDeleteCmd(runtimeContainer)

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

	if injectedTimer.StartCount == 0 {
		t.Fatalf("expected timer Start to be called, got %d", injectedTimer.StartCount)
	}

	if factory.CallCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.CallCount)
	}

	if provisioner.DeleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.DeleteCalls)
	}
}
