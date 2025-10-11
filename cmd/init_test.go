package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmdpkg "github.com/devantler-tech/ksail-go/cmd"
	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

type recordingInitTimer struct {
	startCount int
	stageCount int
}

func (r *recordingInitTimer) Start() {
	r.startCount++
}

func (r *recordingInitTimer) NewStage() {
	r.stageCount++
}

func (r *recordingInitTimer) Stop() {}

func (r *recordingInitTimer) GetTiming() (time.Duration, time.Duration) {
	return time.Millisecond, time.Millisecond
}

func newConfigManagerWithFile(
	t *testing.T,
	writer io.Writer,
	configPath string,
) *ksailconfigmanager.ConfigManager {
	t.Helper()

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	selectors = append(selectors, ksailconfigmanager.StandardSourceDirectoryFieldSelector())

	manager := ksailconfigmanager.NewConfigManager(writer, selectors...)
	manager.Viper.SetConfigFile(configPath)

	return manager
}

func newInitCommand(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "init"}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	return cmd
}

func TestHandleInitRunE_SuccessWithOutputFlag(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	configDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, configDir)

	var buffer bytes.Buffer

	cfgManager := newConfigManagerWithFile(t, &buffer, filepath.Join(configDir, "ksail.yaml"))
	cfgManager.Viper.Set("output", outDir)

	cmd := newInitCommand(t)
	cmd.SetOut(&buffer)
	cmd.SetErr(&buffer)

	tmr := &recordingInitTimer{}
	deps := cmdpkg.InitDeps{Timer: tmr}

	var err error

	err = cmdpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	if tmr.startCount == 0 {
		t.Fatal("expected timer Start to be called")
	}

	if tmr.stageCount == 0 {
		t.Fatal("expected timer NewStage to be called")
	}

	// Normalize temp directory paths for snapshot comparison
	output := strings.ReplaceAll(buffer.String(), configDir, "<config-dir>")

	snaps.MatchSnapshot(t, output)

	_, err = os.Stat(filepath.Join(outDir, "ksail.yaml"))
	if err != nil {
		t.Fatalf("expected ksail.yaml to be scaffolded: %v", err)
	}
}

func TestHandleInitRunE_ReturnsErrorForInvalidConfig(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "ksail.yaml")

	err := os.WriteFile(configPath, []byte(": invalid"), 0o600)
	if err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	cfgManager := newConfigManagerWithFile(t, io.Discard, configPath)
	cmd := newInitCommand(t)

	tmr := &recordingInitTimer{}
	deps := cmdpkg.InitDeps{Timer: tmr}

	err = cmdpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err == nil {
		t.Fatal("expected error from invalid config")
	}

	if !strings.Contains(err.Error(), "failed to load cluster configuration") {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmr.startCount == 0 {
		t.Fatal("expected timer Start to be called even on failure")
	}
}

//nolint:paralleltest // Uses t.Chdir for snapshot setup.
func TestHandleInitRunE_UsesWorkingDirectoryWhenOutputUnset(t *testing.T) {
	workingDir := t.TempDir()
	configDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, configDir)

	configPath := filepath.Join(configDir, "ksail.yaml")

	var buffer bytes.Buffer

	cfgManager := newConfigManagerWithFile(t, &buffer, configPath)

	cmd := newInitCommand(t)
	cmd.SetOut(&buffer)
	cmd.SetErr(&buffer)

	t.Chdir(workingDir)

	tmr := &recordingInitTimer{}
	deps := cmdpkg.InitDeps{Timer: tmr}

	var err error

	err = cmdpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	// Normalize temp directory paths for snapshot comparison
	output := strings.ReplaceAll(buffer.String(), configDir, "<config-dir>")

	snaps.MatchSnapshot(t, output)

	_, err = os.Stat(filepath.Join(workingDir, "ksail.yaml"))
	if err != nil {
		t.Fatalf("expected ksail.yaml in working directory: %v", err)
	}
}
