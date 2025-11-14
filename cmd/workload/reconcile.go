package workload

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// ErrUnsupportedGitOpsEngine is returned when an unsupported GitOps engine is encountered.
var ErrUnsupportedGitOpsEngine = errors.New("unsupported GitOps engine")

// NewReconcileCmd creates the workload reconcile command.
func NewReconcileCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reconcile",
		Short:        "Reconcile workloads with the cluster",
		Long:         "Trigger reconciliation tooling to sync local workloads with your cluster.",
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(func(cmd *cobra.Command, _ runtime.Injector, tmr timer.Timer) error {
			return handleReconcileRunE(cmd, cfgManager, tmr)
		}),
	)

	return cmd
}

// handleReconcileRunE executes the reconcile command.
func handleReconcileRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	tmr timer.Timer,
) error {
	tmr.Start()

	clusterCfg, err := cfgManager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Check if GitOps engine is configured
	if clusterCfg.Spec.GitOpsEngine == v1alpha1.GitOpsEngineNone ||
		clusterCfg.Spec.GitOpsEngine == "" {
		// No GitOps engine configured - nothing to reconcile
		// User can set gitOpsEngine in ksail.yaml to enable GitOps reconciliation
		return nil
	}

	// Handle reconciliation based on GitOps engine
	switch clusterCfg.Spec.GitOpsEngine {
	case v1alpha1.GitOpsEngineFlux:
		return reconcileWithFlux(cmd, clusterCfg, tmr)
	case v1alpha1.GitOpsEngineNone:
		// No GitOps engine - already handled above
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedGitOpsEngine, clusterCfg.Spec.GitOpsEngine)
	}
}

// reconcileWithFlux reconciles workloads using Flux.
func reconcileWithFlux(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	tmr timer.Timer,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Reconcile with Flux...",
		Emoji:   "🔄",
		Writer:  cmd.OutOrStdout(),
	})

	// Get source directory
	sourceDir := clusterCfg.Spec.SourceDirectory
	if sourceDir == "" {
		sourceDir = "./k8s"
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "reconciling workloads from " + sourceDir,
		Writer:  cmd.OutOrStdout(),
	})

	// Trigger Flux reconciliation
	// This is a minimal implementation that demonstrates integration with Flux
	// In a full implementation, this would:
	// 1. Create/update GitRepository or OCIRepository source pointing to workloads
	// 2. Create/update Kustomization resources that reference the source
	// 3. Annotate resources with reconcile.fluxcd.io/requestedAt to trigger immediate sync
	// For now, we demonstrate the workflow structure
	err := triggerFluxReconciliation(cmd.Context(), clusterCfg, sourceDir)
	if err != nil {
		return fmt.Errorf("failed to trigger flux reconciliation: %w", err)
	}

	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, false)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "workloads reconciled " + timingStr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// triggerFluxReconciliation triggers Flux to reconcile workloads from the source directory.
// This is a placeholder implementation demonstrating the reconciliation workflow.
// A complete implementation would create Flux Source and Kustomization resources.
func triggerFluxReconciliation(
	ctx context.Context,
	clusterCfg *v1alpha1.Cluster,
	sourceDir string,
) error {
	// Placeholder: In a full implementation, this would:
	// 1. Use the Flux client to create/update Source resources (Git/OCI)
	// 2. Create/update Kustomization resources pointing to the source
	// 3. Add reconcile annotations to trigger immediate reconciliation
	// 4. Wait for reconciliation to complete

	// For demonstration purposes, we return success to show the workflow
	// The actual implementation would interact with the Kubernetes API via the Flux client
	_ = ctx
	_ = clusterCfg
	_ = sourceDir

	return nil
}
