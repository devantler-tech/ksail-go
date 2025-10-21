package workload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

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

	cmd.RunE = newReconcileCommandRunE(runtimeContainer, cfgManager)

	return cmd
}

// newReconcileCommandRunE creates the RunE handler for workload reconciliation.
func newReconcileCommandRunE(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
) func(*cobra.Command, []string) error {
	return shared.WrapLifecycleHandler(runtimeContainer, cfgManager, handleReconcileRunE)
}

// handleReconcileRunE executes workload reconciliation based on the configured GitOps engine.
func handleReconcileRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
) error {
	clusterCfg := cfgManager.GetConfig()

	// Start timer for reconciliation
	tmr := timer.New()
	tmr.Start()

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Reconcile workloads...",
		Emoji:   "🔄",
		Writer:  cmd.OutOrStdout(),
	})

	// Execute reconciliation based on GitOps engine
	var err error
	switch clusterCfg.Spec.GitOpsEngine {
	case v1alpha1.GitOpsEngineFlux:
		err = reconcileFlux(cmd.Context(), clusterCfg)
	case v1alpha1.GitOpsEngineNone:
		notify.WriteMessage(notify.Message{
			Type:    notify.InfoType,
			Content: "No GitOps engine configured, skipping reconciliation",
			Writer:  cmd.OutOrStdout(),
		})
		return nil
	default:
		return fmt.Errorf("unsupported GitOps engine: %s", clusterCfg.Spec.GitOpsEngine)
	}

	if err != nil {
		return fmt.Errorf("reconciliation failed: %w", err)
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

// reconcileFlux reconciles all Flux resources in the source directory.
func reconcileFlux(ctx context.Context, clusterCfg *v1alpha1.Cluster) error {
	kubeconfig, err := expandKubeconfigPath(clusterCfg.Spec.Connection.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to expand kubeconfig path: %w", err)
	}

	// Reconcile source directory kustomizations if they exist
	sourceDir := clusterCfg.Spec.SourceDirectory
	if sourceDir != "" {
		err = reconcileFluxKustomizations(ctx, kubeconfig, clusterCfg.Spec.Connection.Context, sourceDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// reconcileFluxKustomizations reconciles all kustomizations found in the source directory.
func reconcileFluxKustomizations(ctx context.Context, kubeconfig, kubeContext, sourceDir string) error {
	// Use flux reconcile kustomization to reconcile all kustomizations
	args := []string{
		"reconcile",
		"kustomization",
		"--with-source",
	}

	if kubeconfig != "" {
		args = append(args, fmt.Sprintf("--kubeconfig=%s", kubeconfig))
	}

	if kubeContext != "" {
		args = append(args, fmt.Sprintf("--context=%s", kubeContext))
	}

	// Add timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "flux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reconcile Flux kustomizations: %w (output: %s)", err, string(output))
	}

	return nil
}

// expandKubeconfigPath expands tilde (~) in kubeconfig paths to the user's home directory.
func expandKubeconfigPath(kubeconfig string) (string, error) {
	if len(kubeconfig) == 0 || kubeconfig[0] != '~' {
		return kubeconfig, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, kubeconfig[1:]), nil
}
