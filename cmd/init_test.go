package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	cmdpkg "github.com/devantler-tech/ksail-go/cmd"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
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

func newInitCommand(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "init"}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	return cmd
}

func newConfigManager(
	t *testing.T,
	cmd *cobra.Command,
	writer io.Writer,
) *ksailconfigmanager.ConfigManager {
	t.Helper()

	cmd.SetOut(writer)
	cmd.SetErr(writer)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	selectors = append(selectors, ksailconfigmanager.StandardSourceDirectoryFieldSelector())
	selectors = append(selectors, ksailconfigmanager.DefaultCNIFieldSelector())
	selectors = append(selectors, ksailconfigmanager.DefaultGitOpsEngineFieldSelector())

	manager := ksailconfigmanager.NewCommandConfigManager(cmd, selectors)

	cmd.Flags().StringP("output", "o", "", "Output directory for the project")
	_ = manager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	_ = manager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))

	cmd.Flags().StringSlice("mirror-registry", []string{},
		"Configure mirror registries with format 'host=upstream' (e.g., docker.io=https://registry-1.docker.io).")
	_ = manager.Viper.BindPFlag("mirror-registry", cmd.Flags().Lookup("mirror-registry"))

	return manager
}

func TestHandleInitRunE_SuccessWithOutputFlag(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	var buffer bytes.Buffer

	cmd := newInitCommand(t)
	cfgManager := newConfigManager(t, cmd, &buffer)

	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}
	}

	requireNoError(cmd.Flags().Set("output", outDir))
	requireNoError(cmd.Flags().Set("force", "true"))

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

	snaps.MatchSnapshot(t, buffer.String())

	_, err = os.Stat(filepath.Join(outDir, "ksail.yaml"))
	if err != nil {
		t.Fatalf("expected ksail.yaml to be scaffolded: %v", err)
	}
}

func TestHandleInitRunE_RespectsDistributionFlag(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()

	var buffer bytes.Buffer

	cmd := newInitCommand(t)
	cfgManager := newConfigManager(t, cmd, &buffer)

	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}
	}

	requireNoError(cmd.Flags().Set("output", outDir))
	requireNoError(cmd.Flags().Set("distribution", "K3d"))
	requireNoError(cmd.Flags().Set("distribution-config", "k3d.yaml"))
	requireNoError(cmd.Flags().Set("force", "true"))

	tmr := &recordingInitTimer{}
	deps := cmdpkg.InitDeps{Timer: tmr}

	if err := cmdpkg.HandleInitRunE(cmd, cfgManager, deps); err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "k3d.yaml")); err != nil {
		t.Fatalf("expected k3d.yaml to be scaffolded: %v", err)
	}
}

//nolint:paralleltest // Uses t.Chdir for snapshot setup.
func TestHandleInitRunE_UsesWorkingDirectoryWhenOutputUnset(t *testing.T) {
	workingDir := t.TempDir()
	var buffer bytes.Buffer

	cmd := newInitCommand(t)
	cfgManager := newConfigManager(t, cmd, &buffer)

	t.Chdir(workingDir)

	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}
	}

	requireNoError(cmd.Flags().Set("force", "true"))

	tmr := &recordingInitTimer{}
	deps := cmdpkg.InitDeps{Timer: tmr}

	var err error

	err = cmdpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())

	_, err = os.Stat(filepath.Join(workingDir, "ksail.yaml"))
	if err != nil {
		t.Fatalf("expected ksail.yaml in working directory: %v", err)
	}
}
