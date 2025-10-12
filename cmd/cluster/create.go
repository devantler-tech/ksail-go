package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/repo"
)

// newCreateLifecycleConfig creates the lifecycle configuration for cluster creation.
func newCreateLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🚀",
		TitleContent:       "Create cluster...",
		ActivityContent:    "creating cluster",
		SuccessContent:     "cluster created",
		ErrorMessagePrefix: "failed to create cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Create(ctx, clusterName)
		},
	}
}

// NewCreateCmd wires the cluster create command using the shared runtime container.
func NewCreateCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = newCreateCommandRunE(runtimeContainer, cfgManager)

	return cmd
}

// newCreateCommandRunE creates the RunE handler for cluster creation with CNI installation support.
func newCreateCommandRunE(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
) func(*cobra.Command, []string) error {
	return runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(
			func(cmd *cobra.Command, injector runtime.Injector, tmr timer.Timer) error {
				factory, err := runtime.ResolveClusterProvisionerFactory(injector)
				if err != nil {
					return fmt.Errorf("resolve provisioner factory dependency: %w", err)
				}

				deps := shared.LifecycleDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return handleCreateRunE(cmd, cfgManager, deps)
			},
		),
	)
}

// handleCreateRunE executes cluster creation with CNI installation.
func handleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
) error {
	config := newCreateLifecycleConfig()

	// Reuse the standard lifecycle logic but extend with CNI installation
	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		return fmt.Errorf("cluster creation failed: %w", err)
	}

	// Install CNI if Cilium is configured
	clusterCfg := cfgManager.GetConfig()
	if clusterCfg.Spec.CNI == v1alpha1.CNICilium {
		err = installCiliumCNI(cmd.Context(), clusterCfg)
		if err != nil {
			return fmt.Errorf("failed to install Cilium CNI: %w", err)
		}
	}

	return nil
}

// installCiliumCNI installs Cilium CNI on the cluster.
func installCiliumCNI(ctx context.Context, clusterCfg *v1alpha1.Cluster) error {
	// Determine kubeconfig path
	kubeconfig := clusterCfg.Spec.Connection.Kubeconfig
	if kubeconfig == "" {
		// Use default kubeconfig location
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Read kubeconfig file
	// #nosec G304 - kubeconfig path is from cluster configuration, not user input
	kubeconfigData, err := os.ReadFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig file %s: %w", kubeconfig, err)
	}

	// Create Helm client using kubeconfig
	helmClient, err := helmclient.NewClientFromKubeConf(&helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        "kube-system",
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            false,
			Linting:          false,
			DebugLog:         nil,
			RegistryConfig:   "",
			Output:           nil,
		},
		KubeContext: clusterCfg.Spec.Connection.Context,
		KubeConfig:  kubeconfigData, // Pass actual kubeconfig content
	})
	if err != nil {
		return fmt.Errorf("failed to create Helm client: %w", err)
	}

	// Add Cilium Helm repository
	err = helmClient.AddOrUpdateChartRepo(repo.Entry{
		Name: "cilium",
		URL:  "https://helm.cilium.io/",
	})
	if err != nil {
		return fmt.Errorf("failed to add Cilium Helm repository: %w", err)
	}

	// Determine timeout
	const defaultTimeout = 5

	timeout := defaultTimeout * time.Minute
	if clusterCfg.Spec.Connection.Timeout.Duration > 0 {
		timeout = clusterCfg.Spec.Connection.Timeout.Duration
	}

	// Create Cilium installer
	installer := ciliuminstaller.NewCiliumInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)

	// Install Cilium
	err = installer.Install(ctx)
	if err != nil {
		return fmt.Errorf("cilium installation failed: %w", err)
	}

	return nil
}
