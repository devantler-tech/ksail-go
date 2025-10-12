// Package testutils provides shared helpers for cluster command tests.
package testutils

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/helpers"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// SetupValidWorkingDir creates a valid KSail configuration in a temporary directory and switches to it.
// The returned cleanup function restores the original working directory.
func SetupValidWorkingDir(t *testing.T) func() {
	t.Helper()

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	t.Chdir(tempDir)

	return func() {}
}

// RunValidationErrorTest executes the provided command factory in an empty directory and validates error output.
func RunValidationErrorTest(
	t *testing.T,
	commandFactory func() *cobra.Command,
	expectedSubstrings ...string,
) {
	t.Helper()

	tempDir := t.TempDir()

	t.Chdir(tempDir)

	command := commandFactory()

	var out bytes.Buffer
	command.SetOut(&out)
	command.SetErr(&out)

	if command.RunE == nil {
		t.Fatal("command RunE must not be nil")
	}

	err := command.RunE(command, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	message := err.Error()
	requiredParts := append(
		[]string{
			"failed to load cluster configuration",
			helpers.ErrConfigurationValidationFailed.Error(),
		},
		expectedSubstrings...,
	)

	for _, substring := range requiredParts {
		if !strings.Contains(message, substring) {
			t.Fatalf("expected error message to contain %q, got %q", substring, message)
		}
	}
}

// RecordingTimer is a test double for timer.Timer that records method calls.
type RecordingTimer struct {
	StartCount    int
	NewStageCount int
}

// Start implements timer.Timer.
func (r *RecordingTimer) Start() { r.StartCount++ }

// NewStage implements timer.Timer.
func (r *RecordingTimer) NewStage() { r.NewStageCount++ }

// GetTiming implements timer.Timer.
func (r *RecordingTimer) GetTiming() (time.Duration, time.Duration) {
	return 0, 0
}

// Stop implements timer.Timer.
func (r *RecordingTimer) Stop() {}

// StubFactory is a test double for clusterprovisioner.Factory.
type StubFactory struct {
	Provisioner        clusterprovisioner.ClusterProvisioner
	DistributionConfig any
	Err                error
	CallCount          int
}

// Create implements clusterprovisioner.Factory.
//
//nolint:ireturn // Tests depend on returning the interface type.
func (s *StubFactory) Create(
	_ context.Context,
	_ *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	s.CallCount++
	if s.Err != nil {
		return nil, nil, s.Err
	}

	return s.Provisioner, s.DistributionConfig, nil
}

// StubProvisioner is a test double for clusterprovisioner.ClusterProvisioner.
type StubProvisioner struct {
	CreateErr     error
	CreateCalls   int
	DeleteErr     error
	DeleteCalls   int
	StartErr      error
	StartCalls    int
	StopErr       error
	StopCalls     int
	ReceivedNames []string
}

// Create implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) Create(_ context.Context, name string) error {
	p.CreateCalls++
	p.ReceivedNames = append(p.ReceivedNames, name)

	return p.CreateErr
}

// Delete implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) Delete(_ context.Context, name string) error {
	p.DeleteCalls++
	p.ReceivedNames = append(p.ReceivedNames, name)

	return p.DeleteErr
}

// Start implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) Start(_ context.Context, name string) error {
	p.StartCalls++
	p.ReceivedNames = append(p.ReceivedNames, name)

	return p.StartErr
}

// Stop implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) Stop(_ context.Context, name string) error {
	p.StopCalls++
	p.ReceivedNames = append(p.ReceivedNames, name)

	return p.StopErr
}

// List implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) List(context.Context) ([]string, error) {
	return nil, nil
}

// Exists implements clusterprovisioner.ClusterProvisioner.
func (p *StubProvisioner) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// NewCommand creates a test command with output buffers.
func NewCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	return cmd, &out
}

// CreateConfigManager creates a config manager for testing with a valid ksail config.
func CreateConfigManager(t *testing.T, writer io.Writer) *ksailconfigmanager.ConfigManager {
	t.Helper()

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(writer, selectors...)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	cfgManager.Viper.SetConfigFile(filepath.Join(tempDir, "ksail.yaml"))

	return cfgManager
}

var (
	_ timer.Timer                           = (*RecordingTimer)(nil)
	_ clusterprovisioner.Factory            = (*StubFactory)(nil)
	_ clusterprovisioner.ClusterProvisioner = (*StubProvisioner)(nil)
)
