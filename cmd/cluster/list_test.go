package cluster //nolint:testpackage // Interacts with unexported helpers and shared fixtures.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

func TestHandleListRunE(t *testing.T) {
	t.Parallel()

	t.Run("running clusters", func(t *testing.T) {
		t.Parallel()

		runListRunningClusters(t)
	})

	t.Run("all flag", func(t *testing.T) {
		t.Parallel()

		runListAllFlag(t)
	})

	t.Run("load failure", func(t *testing.T) {
		t.Parallel()

		runListLoadFailure(t)
	})
}

func runListRunningClusters(t *testing.T) {
	t.Helper()

	runListScenario(
		t,
		func(t *testing.T, _ *cobra.Command, manager *configmanager.ConfigManager, _ *bytes.Buffer) {
			t.Helper()
			testutils.SeedValidClusterConfig(manager)
		},
		func(t *testing.T, buffer *bytes.Buffer, err error) {
			t.Helper()

			assertListSuccess(t, buffer, err,
				"Listing running clusters",
				"Distribution filter: Kind",
			)
		},
	)
}

func runListAllFlag(t *testing.T) {
	t.Helper()

	runListScenario(
		t,
		func(t *testing.T, cmd *cobra.Command, manager *configmanager.ConfigManager, _ *bytes.Buffer) {
			t.Helper()

			err := cmd.Flags().Set("all", "true")
			if err != nil {
				t.Fatalf("failed to set all flag: %v", err)
			}

			testutils.SeedValidClusterConfig(manager)
		},
		func(t *testing.T, buffer *bytes.Buffer, err error) {
			t.Helper()

			assertListSuccess(t, buffer, err, "Listing all clusters")
		},
	)
}

func runListLoadFailure(t *testing.T) {
	t.Helper()

	runListScenario(
		t,
		func(t *testing.T, _ *cobra.Command, manager *configmanager.ConfigManager, _ *bytes.Buffer) {
			t.Helper()
			manager.Viper.SetConfigFile(t.TempDir())
		},
		func(t *testing.T, buffer *bytes.Buffer, err error) {
			t.Helper()

			if err == nil {
				t.Fatal("expected error but got nil")
			}

			if !strings.Contains(err.Error(), "failed to read config file") {
				t.Fatalf("expected read config file error, got %v", err)
			}
		},
	)
}

func runListScenario(
	t *testing.T,
	configure func(*testing.T, *cobra.Command, *configmanager.ConfigManager, *bytes.Buffer),
	assert func(*testing.T, *bytes.Buffer, error),
) {
	t.Helper()

	cmd := NewListCmd()
	buffer := configureListCommand(t, cmd)
	manager := configmanager.NewConfigManager(buffer)

	if configure != nil {
		configure(t, cmd, manager, buffer)
	}

	err := HandleListRunE(cmd, manager, nil)
	assert(t, buffer, err)
}

func assertListSuccess(t *testing.T, buffer *bytes.Buffer, err error, expectedMessages ...string) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buffer.String()

	for _, expected := range expectedMessages {
		assertOutputContains(t, output, expected)
	}
}

func configureListCommand(t *testing.T, cmd *cobra.Command) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)

	return buffer
}
