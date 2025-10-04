package workload_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/cmd/workload"
	internaltestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { internaltestutils.RunTestMainWithSnapshotCleanup(m) }

func TestWorkloadHelpSnapshots(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{name: "namespace", args: []string{"workload", "--help"}},
		{name: "reconcile", args: []string{"workload", "reconcile", "--help"}},
		{name: "apply", args: []string{"workload", "apply", "--help"}},
		{name: "install", args: []string{"workload", "install", "--help"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			root := cmd.NewRootCmd("test", "test", "test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(testCase.args)

			err := root.Execute()
			require.NoErrorf(
				t,
				err,
				"expected no error executing %s help",
				strings.Join(testCase.args, " "),
			)

			snaps.MatchSnapshot(t, out.String())
		})
	}
}

//nolint:paralleltest,tparallel // Cannot use t.Parallel() because test changes directories using t.Chdir()
func TestWorkloadCommandsLoadConfigOnly(t *testing.T) {
	handlers := []struct {
		name    string
		handler func(*cobra.Command, *configmanager.ConfigManager, []string) error
	}{
		{name: "reconcile", handler: workload.HandleReconcileRunE},
		{name: "apply", handler: workload.HandleApplyRunE},
		{name: "install", handler: workload.HandleInstallRunE},
	}

	for _, testCase := range handlers {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			manager := newConfigManagerWithDefaults(&out)

			cmd := &cobra.Command{Use: testCase.name}

			err := testCase.handler(cmd, manager, nil)
			require.NoErrorf(t, err, "expected workload %s handler to succeed", testCase.name)

			actual := out.String()
			require.Contains(t, actual, "config loaded")
			require.NotContains(t, actual, "coming soon")
			require.NotContains(t, actual, "ℹ")
		})
	}
}

func TestNewWorkloadCmdRunETriggersHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	command := workload.NewWorkloadCmd()
	command.SetOut(&out)
	command.SetErr(&out)

	err := command.Execute()
	require.NoError(t, err)

	output := out.String()
	if !strings.Contains(output, "Group workload commands under a single namespace") {
		t.Fatalf("expected help output to mention workload namespace details, got %q", output)
	}
}

func TestWorkloadCommandConfiguration(t *testing.T) {
	t.Parallel()

	command := workload.NewWorkloadCmd()

	require.False(t, command.SilenceErrors)
	require.False(t, command.SilenceUsage)
	require.Equal(t, helpers.SuggestionsMinimumDistance, command.SuggestionsMinimumDistance)
}

func newConfigManagerWithDefaults(writer io.Writer) *configmanager.ConfigManager {
	return configmanager.NewConfigManager(
		writer,
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.APIVersion },
			Description:  "API version",
			DefaultValue: v1alpha1.APIVersion,
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Kind },
			Description:  "Resource kind",
			DefaultValue: v1alpha1.Kind,
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to use",
			DefaultValue: v1alpha1.DistributionKind,
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			Description:  "Path to distribution configuration file",
			DefaultValue: "kind.yaml",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context name",
			DefaultValue: "kind-kind",
		},
	)
}
