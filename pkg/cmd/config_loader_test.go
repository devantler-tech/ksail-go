package cmd_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	pkgcmd "github.com/devantler-tech/ksail-go/pkg/cmd"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

type recordingTimer struct {
	started bool
}

func (r *recordingTimer) Start()                                    { r.started = true }
func (r *recordingTimer) NewStage()                                 {}
func (r *recordingTimer) GetTiming() (time.Duration, time.Duration) { return 0, 0 }
func (r *recordingTimer) Stop()                                     {}

const validConfigYAML = `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: sample
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
`

func TestLoadConfigStartsTimer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ksail.yaml")
	writeConfig(t, path, validConfigYAML)

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(path)

	recorder := &recordingTimer{}
	deps := pkgcmd.ConfigLoadDeps{Timer: recorder}

	err := pkgcmd.LoadConfig(cfgManager, deps)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if !recorder.started {
		t.Fatal("expected timer to start")
	}
}

func TestLoadConfigReturnsWrappedError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ksail.yaml")
	writeConfig(t, path, ": invalid yaml")

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(path)

	err := pkgcmd.LoadConfig(cfgManager, pkgcmd.ConfigLoadDeps{})
	if err == nil {
		t.Fatal("expected load error")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest
func TestNewConfigLoaderRunESuccess(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, filepath.Join(dir, "ksail.yaml"), validConfigYAML)

	var captured *recordingTimer

	runtimeContainer := runtime.New(func(injector runtime.Injector) error {
		do.Provide(injector, func(do.Injector) (timer.Timer, error) {
			recorder := &recordingTimer{}
			captured = recorder

			return recorder, nil
		})

		return nil
	})

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	t.Chdir(dir)

	err := pkgcmd.NewConfigLoaderRunE(runtimeContainer)(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	if captured == nil || !captured.started {
		t.Fatal("expected timer to be resolved and started")
	}
}

func TestNewConfigLoaderRunETimerResolutionError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(io.Discard)

	err := pkgcmd.NewConfigLoaderRunE(runtime.New())(cmd, nil)
	if err == nil {
		t.Fatal("expected timer resolution error")
	}

	if !strings.Contains(err.Error(), "resolve timer dependency") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeConfig(t *testing.T, path, contents string) {
	t.Helper()

	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}
}
