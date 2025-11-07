package cluster_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	clusterpkg "github.com/devantler-tech/ksail-go/cmd/cluster"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	cmdtestutils "github.com/devantler-tech/ksail-go/pkg/testutils"
	timermocks "github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

const mirrorRegistryHelp = "Configure mirror registries with format 'host=upstream' " +
	"(e.g., docker.io=https://registry-1.docker.io)."

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
	manager := ksailconfigmanager.NewCommandConfigManager(cmd, clusterpkg.InitFieldSelectors())
	// bind init-local flags like production code
	cmd.Flags().StringP("output", "o", "", "Output directory for the project")
	_ = manager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	_ = manager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))
	cmd.Flags().
		StringSlice("mirror-registry", []string{}, mirrorRegistryHelp)
	_ = manager.Viper.BindPFlag("mirror-registry", cmd.Flags().Lookup("mirror-registry"))

	return manager
}

func TestHandleInitRunE_SuccessWithOutputFlag(t *testing.T) {
	t.Parallel()

	// Using mockery-generated Timer (pkg/ui/timer/mocks.go) so we can set deterministic
	// expectations on timing calls without maintaining a bespoke RecordingTimer helper.

	outDir := t.TempDir()

	var buffer bytes.Buffer

	cmd := newInitCommand(t)
	cfgManager := newConfigManager(t, cmd, &buffer)

	cmdtestutils.SetFlags(t, cmd, map[string]string{
		"output": outDir,
		"force":  "true",
	})

	deps := newInitDeps(t)

	var err error

	err = clusterpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	// Expectations asserted via mock cleanup

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

	cmdtestutils.SetFlags(t, cmd, map[string]string{
		"output":              outDir,
		"distribution":        "K3d",
		"distribution-config": "k3d.yaml",
		"force":               "true",
	})

	deps := newInitDeps(t)

	err := clusterpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	_, err = os.Stat(filepath.Join(outDir, "k3d.yaml"))
	if err != nil {
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

	cmdtestutils.SetFlags(t, cmd, map[string]string{
		"force": "true",
	})

	deps := newInitDeps(t)

	var err error

	err = clusterpkg.HandleInitRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("HandleInitRunE returned error: %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())

	_, err = os.Stat(filepath.Join(workingDir, "ksail.yaml"))
	if err != nil {
		t.Fatalf("expected ksail.yaml in working directory: %v", err)
	}
}

func newInitDeps(t *testing.T) clusterpkg.InitDeps {
	t.Helper()
	tmr := timermocks.NewMockTimer(t)
	tmr.EXPECT().Start().Return()
	tmr.EXPECT().NewStage().Return()
	tmr.EXPECT().GetTiming().Return(time.Millisecond, time.Millisecond)

	return clusterpkg.InitDeps{Timer: tmr}
}
