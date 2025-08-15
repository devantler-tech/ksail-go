// Package cmd provides the command-line interface for KSail.
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
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	configValidator *validators.ConfigValidator
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
	RunE: func(cmd *cobra.Command, _ []string) error {
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

	_, err = factory.ClusterProvisioner(&ksailConfig)
	if err != nil {
		return err
	}

	_, err = factory.ContainerEngineProvisioner(&ksailConfig)
	if err != nil {
		return err
	}

	_, err = factory.ReconciliationTool(&ksailConfig)
	if err != nil {
		return err
	}

	configValidator = validators.NewConfigValidator(&ksailConfig)

	// Load distribution configs for validation
	switch ksailConfig.Spec.Distribution {
	case ksailcluster.DistributionKind:
		kindCfg, err := loader.NewKindConfigLoader().Load()
		if err != nil {
			return err
		}

		configValidator.SetDistributionConfigs(&kindCfg, nil)
	case ksailcluster.DistributionK3d:
		k3dCfg, err := loader.NewK3dConfigLoader().Load()
		if err != nil {
			return err
		}

		configValidator.SetDistributionConfigs(nil, &k3dCfg)
	}

	return nil
}

// LoadKSailConfig loads the KSail configuration.
func LoadKSailConfig() (ksailcluster.Cluster, error) {
	return loader.NewKSailConfigLoader().Load()
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

	for index, line := range lines {
		switch {
		case index < yellowLines:
			printYellow(line)
		case index == yellowLines:
			printBlue(line)
		case index > yellowLines && index < 7:
			printGreenBlueCyan(line)
		case index > 6 && index < len(lines)-2:
			printGreenCyan(line)
		default:
			printBlue(line)
		}
	}
}

func printYellow(line string) {
	_, _ = color.New(color.FgYellow, color.Bold).Println(line)
}

func printBlue(line string) {
	_, _ = color.New(color.FgBlue, color.Bold).Println(line)
}

func printGreenBlueCyan(line string) {
	charThirtyEight := 38
	if len(line) >= charThirtyEight {
		_, _ = color.New(color.FgGreen, color.Bold).Print(line[:32])
		_, _ = color.New(color.FgCyan).Print(line[32:37])
		_, _ = color.New(color.FgBlue, color.Bold).Print(line[37:charThirtyEight])
		_, _ = color.New(color.FgCyan).Println(line[38:])
	} else {
		_, _ = color.New(color.FgGreen, color.Bold).Println(line)
	}
}

const greenCyanSplitIndex = 32

func printGreenCyan(line string) {
	if len(line) >= greenCyanSplitIndex {
		_, _ = color.New(color.FgGreen, color.Bold).Print(line[:greenCyanSplitIndex])
		_, _ = color.New(color.FgCyan).Println(line[greenCyanSplitIndex:])
	} else {
		_, _ = color.New(color.FgGreen, color.Bold).Println(line)
	}
}
