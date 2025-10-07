package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

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

var (
	errFactoryBoom  = errors.New("factory boom")
	errCreateFailed = errors.New("create failed")
)

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

type stubProvisioner struct {
	createErr     error
	createCalls   int
	receivedNames []string
}

func (p *stubProvisioner) Create(_ context.Context, name string) error {
	p.createCalls++
	p.receivedNames = append(p.receivedNames, name)

	return p.createErr
}

func (p *stubProvisioner) Delete(context.Context, string) error { return nil }
func (p *stubProvisioner) Start(context.Context, string) error  { return nil }
func (p *stubProvisioner) Stop(context.Context, string) error   { return nil }
func (p *stubProvisioner) List(context.Context) ([]string, error) {
	return nil, nil
}

func (p *stubProvisioner) Exists(context.Context, string) (bool, error) {
	return false, nil
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

func TestHandleCreateRunE_LoadConfigFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)

	timerStub := &recordingTimer{}
	factoryCalled := 0
	failingFactory := &stubFactory{
		err: nil,
	}

	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(badPath, []byte(": invalid"), 0o600)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badPath)

	deps := CreateDeps{Timer: timerStub, Factory: failingFactory}

	err = HandleCreateRunE(cmd, cfgManager, deps)
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

	if factoryCalled != 0 {
		t.Fatalf("expected factory not to be invoked, got %d", factoryCalled)
	}
}

func TestHandleCreateRunE_FactoryFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{err: errFactoryBoom}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
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

func TestHandleCreateRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
	if !errors.Is(err, errMissingClusterProvisioner) {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

func TestHandleCreateRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	factory := &stubFactory{
		provisioner:        &stubProvisioner{},
		distributionConfig: struct{}{},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
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

func TestHandleCreateRunE_ReturnsErrorWhenProvisionerCreateFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCreateCommand(t)
	timerStub := &recordingTimer{}
	provisioner := &stubProvisioner{createErr: errCreateFailed}
	factory := &stubFactory{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, io.Discard)

	err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected provisioner create error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create cluster") {
		t.Fatalf("expected create failure message, got %q", err)
	}

	if provisioner.createCalls != 1 {
		t.Fatalf("expected provisioner Create to be called once, got %d", provisioner.createCalls)
	}
}

func TestHandleCreateRunE_Success(t *testing.T) {
	t.Parallel()

	cmd, out := newCreateCommand(t)
	timerStub := &recordingTimer{}
	provisioner := &stubProvisioner{}
	factory := &stubFactory{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createConfigManager(t, out)

	err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
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

	if provisioner.createCalls != 1 {
		t.Fatalf("expected provisioner Create to be called once, got %d", provisioner.createCalls)
	}

	output := out.String()
	if !strings.Contains(output, "Create cluster...") {
		t.Fatalf("expected provisioning title in output, got %q", output)
	}

	if !strings.Contains(output, "cluster created") {
		t.Fatalf("expected success message in output, got %q", output)
	}
}

func TestNewCreateCmd_RunESuccess(t *testing.T) {
	var injectedTimer *recordingTimer

	provisioner := &stubProvisioner{}
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

	cmd := NewCreateCmd(runtimeContainer)

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

	if provisioner.createCalls != 1 {
		t.Fatalf("expected provisioner Create to be called once, got %d", provisioner.createCalls)
	}
}
