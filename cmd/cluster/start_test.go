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

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errStartFailed = errors.New("start failed")

type stubProvisionerForStart struct {
	startErr      error
	startCalls    int
	receivedNames []string
}

func (p *stubProvisionerForStart) Start(_ context.Context, name string) error {
	p.startCalls++
	p.receivedNames = append(p.receivedNames, name)

	return p.startErr
}

func (p *stubProvisionerForStart) Create(context.Context, string) error   { return nil }
func (p *stubProvisionerForStart) Delete(context.Context, string) error   { return nil }
func (p *stubProvisionerForStart) Stop(context.Context, string) error     { return nil }
func (p *stubProvisionerForStart) List(context.Context) ([]string, error) { return nil, nil }

func (p *stubProvisionerForStart) Exists(
	context.Context,
	string,
) (bool, error) {
	return false, nil
}

func TestHandleStartRunE_LoadConfigFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)

	timerStub := &recordingTimer{}
	failingFactory := &stubFactory{err: nil}

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(badPath, []byte(": invalid"), 0o600)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	deps := StartDeps{Timer: timerStub, Factory: failingFactory}

	err = HandleStartRunE(cmd, cfgManager, deps)
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

func TestHandleStartRunE_FactoryFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{err: errFactoryBoom}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleStartRunE(cmd, cfgManager, StartDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("expected factory failure message, got %q", err)
	}

	if timerStub.newStageCount != 1 {
		t.Fatalf(
			"expected timer NewStage to be called before factory, got %d",
			timerStub.newStageCount,
		)
	}

	if factory.callCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}
}

func TestHandleStartRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleStartRunE(cmd, cfgManager, StartDeps{Timer: timerStub, Factory: factory})
	if !errors.Is(err, errMissingClusterProvisionerStart) {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

func TestHandleStartRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{
		provisioner:        &stubProvisionerForStart{},
		distributionConfig: struct{}{},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleStartRunE(cmd, cfgManager, StartDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected cluster name error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get cluster name") {
		t.Fatalf("expected cluster name failure message, got %q", err)
	}

	if factory.callCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}
}

func TestHandleStartRunE_ReturnsErrorWhenProvisionerStartFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	provisioner := &stubProvisionerForStart{startErr: errStartFailed}
	factory := &stubFactory{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleStartRunE(cmd, cfgManager, StartDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected provisioner start error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to start cluster") {
		t.Fatalf("expected start failure message, got %q", err)
	}

	if provisioner.startCalls != 1 {
		t.Fatalf("expected provisioner Start to be called once, got %d", provisioner.startCalls)
	}
}

func TestHandleStartRunE_Success(t *testing.T) {
	t.Parallel()

	cmd, out := newCreateCommand(t)
	timerStub := &recordingTimer{}
	provisioner := &stubProvisionerForStart{}
	factory := &stubFactory{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, out)

	err := HandleStartRunE(cmd, cfgManager, StartDeps{Timer: timerStub, Factory: factory})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if timerStub.startCount != 1 || timerStub.newStageCount != 1 {
		t.Fatalf(
			"expected timer Start/NewStage to be called once, got %d/%d",
			timerStub.startCount,
			timerStub.newStageCount,
		)
	}

	if provisioner.startCalls != 1 {
		t.Fatalf("expected provisioner Start to be called once, got %d", provisioner.startCalls)
	}

	output := out.String()
	if !strings.Contains(output, "Start cluster...") {
		t.Fatalf("expected start title in output, got %q", output)
	}

	if !strings.Contains(output, "cluster started") {
		t.Fatalf("expected success message in output, got %q", output)
	}
}

//nolint:paralleltest
func TestNewStartCmd_RunESuccess(t *testing.T) {
	var injectedTimer *recordingTimer

	provisioner := &stubProvisionerForStart{}
	factory := &stubFactory{
		provisioner:        provisioner,
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

	cmd := NewStartCmd(runtimeContainer)

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

	if injectedTimer.startCount == 0 {
		t.Fatalf("expected timer Start to be called, got %d", injectedTimer.startCount)
	}

	if factory.callCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}

	if provisioner.startCalls != 1 {
		t.Fatalf("expected provisioner Start to be called once, got %d", provisioner.startCalls)
	}
}
