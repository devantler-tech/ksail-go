package workload

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	helminstaller "github.com/devantler-tech/ksail-go/pkg/installer/helm"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
)

// installFlags holds the flags for the install command.
type installFlags struct {
	releaseName string
	chartName   string
	namespace   string
	version     string
	valuesFile  string
	timeout     time.Duration
}

// NewInstallCmd creates the workload install command.
func NewInstallCmd(_ *runtime.Runtime) *cobra.Command {
	flags := &installFlags{}

	cmd := &cobra.Command{
		Use:   "install [RELEASE_NAME] [CHART]",
		Short: "Install Helm charts",
		Long:  "Install Helm charts to provision workloads through KSail.",
		Args:  cobra.ExactArgs(2),
		Example: `  # Install a chart from a repository
  ksail workload install my-release stable/nginx-ingress

  # Install a chart with a specific version
  ksail workload install my-release stable/nginx-ingress --version 1.2.3

  # Install a chart with custom values
  ksail workload install my-release stable/nginx-ingress --values values.yaml

  # Install an OCI chart
  ksail workload install my-release oci://registry/repo/chart`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.releaseName = args[0]
			flags.chartName = args[1]

			return runInstall(cmd.OutOrStdout(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "default", "Namespace to install the chart into")
	cmd.Flags().StringVar(&flags.version, "version", "", "Version of the chart to install")
	cmd.Flags().StringVar(&flags.valuesFile, "values", "", "Path to values file")
	cmd.Flags().DurationVar(&flags.timeout, "timeout", 5*time.Minute, "Timeout for the install operation")

	return cmd
}

func runInstall(out io.Writer, flags *installFlags) error {
	tmr := timer.New()
	tmr.Start()

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Installing Helm chart...",
		Emoji:   "📦",
		Writer:  out,
	})

	// Get kubeconfig path
	kubeconfigPath := getKubeconfigPathSilently()

	// Read values file if provided
	valuesYaml := ""
	if flags.valuesFile != "" {
		content, err := os.ReadFile(flags.valuesFile)
		if err != nil {
			return fmt.Errorf("failed to read values file: %w", err)
		}
		valuesYaml = string(content)
	}

	// Read kubeconfig file
	kubeconfigBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	// Create helm client using KubeConfClientOptions
	client, err := helmclient.NewClientFromKubeConf(&helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace: flags.namespace,
		},
		KubeConfig: kubeconfigBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to create helm client: %w", err)
	}

	// Create installer
	installer := helminstaller.NewHelmInstaller(
		client,
		flags.releaseName,
		flags.chartName,
		flags.namespace,
		flags.version,
		valuesYaml,
		flags.timeout,
	)

	// Install chart
	ctx := context.Background()
	err = installer.Install(ctx)
	if err != nil {
		return fmt.Errorf("failed to install chart: %w", err)
	}

	// Report success
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "Helm chart installed successfully",
		Timer:   tmr,
		Writer:  out,
	})

	return nil
}
