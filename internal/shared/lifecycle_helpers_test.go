package shared_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/spf13/cobra"
)

var errFactoryError = errors.New("factory error")

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

// TestHandleLifecycleRunE_FactoryError validates that an error from Factory.Create is wrapped correctly.
func TestHandleLifecycleRunE_FactoryError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	path := filepath.Join(tempDir, "ksail.yaml")
	if err := os.WriteFile(path, []byte(validClusterConfigYAML), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(path)

	timer := &lifecycleTimer{}
	// Stub factory implementing only Create to force error path after config load.
	factory := &errorFactory{err: errFactoryError}
	deps := shared.LifecycleDeps{Timer: timer, Factory: factory}
	config := shared.LifecycleConfig{}
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err == nil {
		t.Fatal("expected factory error")
	}

	if !strings.Contains(err.Error(), "failed to resolve cluster provisioner") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !timer.started {
		t.Error("expected timer to be started")
	}

	if timer.newStageCalls != 1 {
		t.Errorf("expected newStageCalls=1, got %d", timer.newStageCalls)
	}
}

// errorFactory satisfies clusterprovisioner.Factory for testing.
type errorFactory struct{ err error }

func (e *errorFactory) Create(
	context.Context,
	*v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	return nil, nil, e.err
}

// The following tests referencing removed helpers/new opts have been omitted during migration.

const validClusterConfigYAML = `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: sample
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
`
