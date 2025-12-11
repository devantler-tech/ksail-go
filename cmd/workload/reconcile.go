package workload

import (
	"errors"
	"fmt"
	"strings"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/devantler-tech/ksail-go/pkg/workload/oci"
	"github.com/spf13/cobra"
)

const defaultArtifactTag = "latest"

var errLocalRegistryRequired = errors.New("local registry must be enabled to reconcile workloads")

// NewReconcileCmd creates the workload reconcile command.
//
//nolint:funlen // Cobra command RunE functions typically combine setup, validation, and execution
func NewReconcileCmd(_ *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reconcile",
		Short:        "Reconcile workloads with the cluster",
		Long:         "Trigger reconciliation tooling to sync local workloads with your cluster.",
		SilenceUsage: true,
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		tmr := timer.New()
		tmr.Start()

		fieldSelectors := ksailconfigmanager.DefaultClusterFieldSelectors()
		cfgManager := ksailconfigmanager.NewCommandConfigManager(cmd, fieldSelectors)

		clusterCfg, err := cfgManager.LoadConfig(tmr)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if clusterCfg.Spec.LocalRegistry != v1alpha1.LocalRegistryEnabled {
			return errLocalRegistryRequired
		}

		sourceDir := clusterCfg.Spec.SourceDirectory
		if strings.TrimSpace(sourceDir) == "" {
			sourceDir = v1alpha1.DefaultSourceDirectory
		}

		repoName := sourceDir
		artifactVersion := defaultArtifactTag

		registryPort := clusterCfg.Spec.Options.LocalRegistry.HostPort
		if registryPort == 0 {
			registryPort = v1alpha1.DefaultLocalRegistryPort
		}

		builder := oci.NewWorkloadArtifactBuilder()

		notify.WriteMessage(notify.Message{
			Type:    notify.TitleType,
			Emoji:   "📦",
			Content: "Build workload artifact...",
			Writer:  cmd.OutOrStdout(),
		})

		tmr.NewStage()

		_, err = builder.Build(cmd.Context(), oci.BuildOptions{
			Name:             repoName,
			SourcePath:       sourceDir,
			RegistryEndpoint: fmt.Sprintf("localhost:%d", registryPort),
			Repository:       repoName,
			Version:          artifactVersion,
		})
		if err != nil {
			return fmt.Errorf("build workload artifact: %w", err)
		}

		total, stage := tmr.GetTiming()
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "artifact pushed " + notify.FormatTiming(total, stage, true),
			Writer:  cmd.OutOrStdout(),
		})

		return nil
	}

	return cmd
}
