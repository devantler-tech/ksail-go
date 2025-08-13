package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/devantler-tech/ksail-go/cmd/inputs"
	factory "github.com/devantler-tech/ksail-go/internal/factories"
	"github.com/devantler-tech/ksail-go/internal/loader"
	"github.com/devantler-tech/ksail-go/internal/ui/notify"
	"github.com/devantler-tech/ksail-go/internal/validators"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	reconciliationtoolbootstrapper "github.com/devantler-tech/ksail-go/pkg/bootstrapper/reconciliation_tool"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
	"github.com/spf13/cobra"
)

var (
	ksailConfig                    ksailcluster.Cluster
	clusterProvisioner             clusterprovisioner.ClusterProvisioner
	containerEngineProvisioner     containerengineprovisioner.ContainerEngineProvisioner
	reconciliationToolBootstrapper reconciliationtoolbootstrapper.Bootstrapper
	configValidator                *validators.ConfigValidator
)

//go:embed assets/ascii-art.txt
var asciiArt string

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "ksail",
	Short: "SDK for operating and managing K8s clusters and workloads",
	Long: `KSail is an SDK for operating and managing Kubernetes clusters and workloads.

  Create ephemeral clusters for development and CI purposes, deploy and update workloads, test and validate behavior — all through one concise, declarative interface. Stop stitching together a dozen CLIs; KSail gives you a consistent UX built on the tools you already trust.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRoot(cmd)
	},
}

// SetVersionInfo sets the version string displayed by the root command.
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		notify.Errorf("%s", err)
		os.Exit(1)
	}
}

// InitServices initializes the services required by the CLI.
func InitServices() error {
	ksailConfig, err := loader.NewKSailConfigLoader().Load()
	if err != nil {
		return err
	}
	inputs.SetInputsOrFallback(&ksailConfig)
	clusterProvisioner, err = factory.ClusterProvisioner(&ksailConfig)
	if err != nil {
		return err
	}
	containerEngineProvisioner, err = factory.ContainerEngineProvisioner(&ksailConfig)
	if err != nil {
		return err
	}
	reconciliationToolBootstrapper, err = factory.ReconciliationTool(&ksailConfig)
	if err != nil {
		return err
	}
	configValidator = validators.NewConfigValidator(&ksailConfig)
	return nil
}

// --- internals ---

// handleRoot handles the root command.
func handleRoot(cmd *cobra.Command) error {
	printASCIIArt()
	return cmd.Help()
}

func printASCIIArt() {
	lines := strings.Split(asciiArt, "\n")
	for i, line := range lines {
		if i < 4 {
			fmt.Println("\x1b[1;33m" + line + "\x1b[0m")
		} else if i == 4 {
			fmt.Println("\x1b[1;34m" + line + "\x1b[0m")
		} else if i > 4 && i < 7 {
			// Add bounds checks to avoid panics if ascii-art changes
			if len(line) >= 38 {
				fmt.Print("\x1b[1;32m" + line[:32] + "\x1b[0m")
				fmt.Print("\x1B[36m" + line[32:37] + "\x1b[0m")
				fmt.Print("\x1b[1;34m" + line[37:38] + "\x1b[0m")
				fmt.Println("\x1B[36m" + line[38:] + "\x1b[0m")
			} else {
				fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
			}
		} else if i > 6 && i < len(lines)-2 {
			if len(line) >= 32 {
				fmt.Print("\x1b[1;32m" + line[:32] + "\x1b[0m")
				fmt.Println("\x1B[36m" + line[32:] + "\x1b[0m")
			} else {
				fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
			}
		} else {
			fmt.Println("\x1b[1;34m" + line + "\x1b[0m")
		}
	}
}
