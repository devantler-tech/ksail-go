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
	errFactoryBoomDelete = errors.New("factory boom")
	errDeleteFailed      = errors.New("delete failed")
)

type recordingTimerDelete struct {
	startCount    int
	newStageCount int
}

func (r *recordingTimerDelete) Start()    { r.startCount++ }
func (r *recordingTimerDelete) NewStage() { r.newStageCount++ }
func (r *recordingTimerDelete) GetTiming() (time.Duration, time.Duration) {
	return 0, 0
}
func (r *recordingTimerDelete) Stop() {}

type stubFactoryDelete struct {
	provisioner        clusterprovisioner.ClusterProvisioner
	distributionConfig any
	err                error
	callCount          int
}

//nolint:ireturn // Tests depend on returning the interface type.
func (s *stubFactoryDelete) Create(
	_ context.Context,
	_ *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	s.callCount++
	if s.err != nil {
		return nil, nil, s.err
	}

	return s.provisioner, s.distributionConfig, nil
}

type stubDeleteProvisioner struct {
	deleteErr     error
	deleteCalls   int
	receivedNames []string
}

func (p *stubDeleteProvisioner) Create(context.Context, string) error { return nil }
func (p *stubDeleteProvisioner) Delete(_ context.Context, name string) error {
	p.deleteCalls++
	p.receivedNames = append(p.receivedNames, name)

	return p.deleteErr
}
func (p *stubDeleteProvisioner) Start(context.Context, string) error { return nil }
func (p *stubDeleteProvisioner) Stop(context.Context, string) error  { return nil }
func (p *stubDeleteProvisioner) List(context.Context) ([]string, error) {
	return nil, nil
}

func (p *stubDeleteProvisioner) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func newDeleteCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	return cmd, &out
}

func createDeleteConfigManager(t *testing.T, writer io.Writer) *ksailconfigmanager.ConfigManager {
	t.Helper()

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(writer, selectors...)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	cfgManager.Viper.SetConfigFile(filepath.Join(tempDir, "ksail.yaml"))

	return cfgManager
}

func TestHandleDeleteRunE_LoadConfigFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newDeleteCommand(t)

	timerStub := &recordingTimerDelete{}
	factoryCalled := 0
	failingFactory := &stubFactoryDelete{
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

	deps := DeleteDeps{Timer: timerStub, Factory: failingFactory}

	err = HandleDeleteRunE(cmd, cfgManager, deps)
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

func TestHandleDeleteRunE_FactoryFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newDeleteCommand(t)
	timerStub := &recordingTimerDelete{}
	factory := &stubFactoryDelete{err: errFactoryBoomDelete}
	cfgManager := createDeleteConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
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

func TestHandleDeleteRunE_ReturnsErrorWhenProvisionerIsNil(t *testing.T) {
	t.Parallel()

	cmd, _ := newDeleteCommand(t)
	timerStub := &recordingTimerDelete{}
	factory := &stubFactoryDelete{}
	cfgManager := createDeleteConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if !errors.Is(err, errMissingClusterProvisionerForDelete) {
		t.Fatalf("expected missing provisioner error, got %v", err)
	}
}

func TestHandleDeleteRunE_ReturnsErrorWhenClusterNameFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newDeleteCommand(t)
	timerStub := &recordingTimerDelete{}
	factory := &stubFactoryDelete{
		provisioner:        &stubDeleteProvisioner{},
		distributionConfig: struct{}{},
	}
	cfgManager := createDeleteConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
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

func TestHandleDeleteRunE_ReturnsErrorWhenProvisionerDeleteFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newDeleteCommand(t)
	timerStub := &recordingTimerDelete{}
	provisioner := &stubDeleteProvisioner{deleteErr: errDeleteFailed}
	factory := &stubFactoryDelete{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createDeleteConfigManager(t, io.Discard)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
	if err == nil {
		t.Fatal("expected provisioner delete error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete cluster") {
		t.Fatalf("expected delete failure message, got %q", err)
	}

	if provisioner.deleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.deleteCalls)
	}
}

func TestHandleDeleteRunE_Success(t *testing.T) {
	t.Parallel()

	cmd, out := newDeleteCommand(t)
	timerStub := &recordingTimerDelete{}
	provisioner := &stubDeleteProvisioner{}
	factory := &stubFactoryDelete{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := createDeleteConfigManager(t, out)

	err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
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

	if provisioner.deleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.deleteCalls)
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
	var injectedTimer *recordingTimerDelete

	provisioner := &stubDeleteProvisioner{}
	factory := &stubFactoryDelete{
		provisioner:        provisioner,
		distributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}

	runtimeContainer := runtime.New(
		func(i runtime.Injector) error {
			do.Provide(i, func(runtime.Injector) (timer.Timer, error) {
				injectedTimer = &recordingTimerDelete{}

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

	if injectedTimer.startCount == 0 {
		t.Fatalf("expected timer Start to be called, got %d", injectedTimer.startCount)
	}

	if factory.callCount != 1 {
		t.Fatalf("expected factory Create to be called once, got %d", factory.callCount)
	}

	if provisioner.deleteCalls != 1 {
		t.Fatalf("expected provisioner Delete to be called once, got %d", provisioner.deleteCalls)
	}
}
