package shared

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

type recordingTimer struct {
	started bool
}

func (r *recordingTimer) Start()                                    { r.started = true }
func (r *recordingTimer) NewStage()                                 {}
func (r *recordingTimer) GetTiming() (time.Duration, time.Duration) { return 0, 0 }
func (r *recordingTimer) Stop()                                     {}

const validConfigYAML = "apiVersion: ksail.dev/v1alpha1\nkind: Cluster\nmetadata:\n  name: sample\nspec:\n  distribution: Kind\n  distributionConfig: kind.yaml\n  sourceDirectory: k8s\n"

func TestLoadConfigStartsTimer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ksail.yaml")
	writeConfig(t, path, validConfigYAML)

	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(path)

	recorder := &recordingTimer{}
	deps := ConfigLoadDeps{Timer: recorder}

	if err := LoadConfig(cfgManager, deps); err != nil {
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

	err := LoadConfig(cfgManager, ConfigLoadDeps{})
	if err == nil {
		t.Fatal("expected load error")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewConfigLoaderRunESuccess(t *testing.T) {
	t.Parallel()

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

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	if err := NewConfigLoaderRunE(runtimeContainer)(cmd, nil); err != nil {
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

	err := NewConfigLoaderRunE(runtime.New())(cmd, nil)
	if err == nil {
		t.Fatal("expected timer resolution error")
	}

	if !strings.Contains(err.Error(), "resolve timer dependency") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeConfig(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
}
