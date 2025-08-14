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

// rootCmd represents the root command.
var rootCmd = &cobra.Command{
	Use:   "ksail",
	Short: "SDK for operating and managing K8s clusters and workloads",
	Long: `KSail helps you easily create, manage, and test local Kubernetes clusters and workloads ` +
		`from one simple command line tool.`,
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
	const yellowLines = 4

	lines := strings.Split(asciiArt, "\n")

	for i, line := range lines {
		switch {
		case i < yellowLines:
			printYellow(line)
		case i == yellowLines:
			printBlue(line)
		case i > yellowLines && i < 7:
			printGreenBlueCyan(line)
		case i > 6 && i < len(lines)-2:
			printGreenCyan(line)
		default:
			printBlue(line)
		}
	}
}

func printYellow(line string) {
	fmt.Println("\x1b[1;33m" + line + "\x1b[0m")
}

func printBlue(line string) {
	fmt.Println("\x1b[1;34m" + line + "\x1b[0m")
}

func printGreenBlueCyan(line string) {
	charThirtyEight := 38
	if len(line) >= charThirtyEight {
		fmt.Print("\x1b[1;32m" + line[:32] + "\x1b[0m")
		fmt.Print("\x1B[36m" + line[32:37] + "\x1b[0m")
		fmt.Print("\x1b[1;34m" + line[37:charThirtyEight] + "\x1b[0m")
		fmt.Println("\x1B[36m" + line[38:] + "\x1b[0m")
	} else {
		fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
	}
}

const greenCyanSplitIndex = 32

func printGreenCyan(line string) {
	if len(line) >= greenCyanSplitIndex {
		fmt.Print("\x1b[1;32m" + line[:greenCyanSplitIndex] + "\x1b[0m")
		fmt.Println("\x1B[36m" + line[greenCyanSplitIndex:] + "\x1b[0m")
	} else {
		fmt.Println("\x1b[1;32m" + line + "\x1b[0m")
	}
}

// clusterOperation performs a common cluster operation (start/stop) with shared validation logic.
func clusterOperation(actionMsg, verbMsg, pastMsg string, operation func(clusterprovisioner.ClusterProvisioner, string) error) error {
	fmt.Println()

	provisioner, err := factory.ClusterProvisioner(&ksailConfig)
	if err != nil {
		return err
	}

	containerEngineProvisioner, err := factory.ContainerEngineProvisioner(&ksailConfig)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%s '%s'\n", actionMsg, ksailConfig.Metadata.Name)
	fmt.Printf("► checking '%s' is ready\n", ksailConfig.Spec.ContainerEngine)

	ready, err := containerEngineProvisioner.CheckReady()
	if err != nil || !ready {
		return fmt.Errorf("container engine '%s' is not ready: %v", ksailConfig.Spec.ContainerEngine, err)
	}

	fmt.Printf("✔ '%s' is ready\n", ksailConfig.Spec.ContainerEngine)
	fmt.Printf("► %s '%s'\n", verbMsg, ksailConfig.Metadata.Name)

	exists, err := provisioner.Exists(ksailConfig.Metadata.Name)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("✔ '%s' not found\n", ksailConfig.Metadata.Name)
		return nil
	}

	if err := operation(provisioner, ksailConfig.Metadata.Name); err != nil {
		return err
	}

	fmt.Printf("✔ '%s' %s\n", ksailConfig.Metadata.Name, pastMsg)
	return nil
}
