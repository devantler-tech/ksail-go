package cluster

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		HandleStatusRunE,
		configmanager.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			Description:  "Path to kubeconfig file",
			DefaultValue: "~/.kube/config",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for status check operations",
			DefaultValue: metav1.Duration{Duration: defaultStatusTimeout},
		},
	)
}

// HandleStatusRunE handles the status command.
// Exported for testing purposes.
func HandleStatusRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load and validate cluster configuration
	cluster, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Get context, using background if not available
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Check cluster status
	status, err := checkClusterStatus(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to check cluster status: %w", err)
	}

	// Display status with timing
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "Cluster status: %s",
		Args:    []any{status},
		Timer:   tmr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// checkClusterStatus checks the status of the cluster.
func checkClusterStatus(ctx context.Context, cluster *v1alpha1.Cluster) (string, error) {
	const apiCheckTimeout = 5 * time.Second

	// Get kubeconfig path, use default if not set
	kubeconfigPath := cluster.Spec.Connection.Kubeconfig
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}

	// Expand home directory in path
	if strings.HasPrefix(kubeconfigPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfigPath = strings.Replace(kubeconfigPath, "~", home, 1)
		}
	}

	// Load kubeconfig with context
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	// Use the context from the cluster config if specified
	if cluster.Spec.Connection.Context != "" {
		configOverrides.CurrentContext = cluster.Spec.Connection.Context
	}

	// Try to connect to the Kubernetes API
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		// If we can't load kubeconfig, check if cluster exists via provisioner
		return checkClusterExistence(ctx, cluster)
	}

	// Try to create a clientset and make a simple API call
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return checkClusterExistence(ctx, cluster)
	}

	// Use a short timeout for the API call
	checkCtx, cancel := context.WithTimeout(ctx, apiCheckTimeout)
	defer cancel()

	// Try to get server version - lightweight check
	_, err = clientset.Discovery().ServerVersion()
	if err != nil {
		// API is not accessible, check if cluster exists
		return checkClusterExistence(checkCtx, cluster)
	}

	return "Running", nil
}

// checkClusterExistence checks if the cluster exists using the provisioner.
func checkClusterExistence(ctx context.Context, cluster *v1alpha1.Cluster) (string, error) {
	// Only check existence for distributions we can manage (Kind, K3d)
	if cluster.Spec.Distribution != v1alpha1.DistributionKind &&
		cluster.Spec.Distribution != v1alpha1.DistributionK3d {
		return "Unknown", nil
	}

	// Create provisioner
	provisioner, clusterName, err := clusterprovisioner.CreateClusterProvisioner(
		ctx,
		cluster.Spec.Distribution,
		cluster.Spec.DistributionConfig,
		cluster.Spec.Connection.Kubeconfig,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create provisioner: %w", err)
	}

	// Check if cluster exists
	exists, err := provisioner.Exists(ctx, clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to check cluster existence: %w", err)
	}

	if !exists {
		return "Not Found", nil
	}

	// Cluster exists but API is not responding
	return "Stopped", nil
}
