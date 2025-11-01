// Package gen provides the gen command namespace for generating Kubernetes resources.
package gen

import (
	"errors"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

var (
	// ErrResourceCommandNotFound is returned when a kubectl create subcommand is not found.
	ErrResourceCommandNotFound = errors.New("kubectl create command not found for resource type")
	// ErrNoRunFunction is returned when a kubectl command has neither RunE nor Run function.
	ErrNoRunFunction = errors.New("no run function found for kubectl create command")
)

// NewGenCmd creates and returns the gen command group namespace.
func NewGenCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate Kubernetes resource manifests",
		Long: "Generate Kubernetes resource manifests using kubectl create with --dry-run=client -o yaml. " +
			"The generated YAML is printed to stdout and can be redirected to a file using shell redirection (> file.yaml).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmd.AddCommand(NewClusterRoleCmd(runtimeContainer))
	cmd.AddCommand(NewClusterRoleBindingCmd(runtimeContainer))
	cmd.AddCommand(NewConfigMapCmd(runtimeContainer))
	cmd.AddCommand(NewCronJobCmd(runtimeContainer))
	cmd.AddCommand(NewDeploymentCmd(runtimeContainer))
	cmd.AddCommand(NewIngressCmd(runtimeContainer))
	cmd.AddCommand(NewJobCmd(runtimeContainer))
	cmd.AddCommand(NewNamespaceCmd(runtimeContainer))
	cmd.AddCommand(NewPodDisruptionBudgetCmd(runtimeContainer))
	cmd.AddCommand(NewPriorityClassCmd(runtimeContainer))
	cmd.AddCommand(NewQuotaCmd(runtimeContainer))
	cmd.AddCommand(NewRoleCmd(runtimeContainer))
	cmd.AddCommand(NewRoleBindingCmd(runtimeContainer))
	cmd.AddCommand(NewSecretCmd(runtimeContainer))
	cmd.AddCommand(NewServiceCmd(runtimeContainer))
	cmd.AddCommand(NewServiceAccountCmd(runtimeContainer))
	cmd.AddCommand(NewTokenCmd(runtimeContainer))

	return cmd
}

// createGenCommand creates a gen subcommand that wraps kubectl create with forced --dry-run=client -o yaml.
func createGenCommand(_ *runtime.Runtime, resourceType string) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create temporary kubectl client to get the resource command
	tempIOStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	tempClient := kubectl.NewClient(tempIOStreams)
	tempCreateCmd := tempClient.CreateCreateCommand(kubeconfigPath)

	// Find the subcommand for this resource type
	var resourceCmd *cobra.Command

	for _, subCmd := range tempCreateCmd.Commands() {
		if subCmd.Name() == resourceType {
			resourceCmd = subCmd

			break
		}
	}

	if resourceCmd == nil {
		panic(fmt.Sprintf("kubectl create %s command not found", resourceType))
	}

	// Create a wrapper command
	wrapperCmd := &cobra.Command{
		Use:          resourceCmd.Use,
		Short:        resourceCmd.Short,
		Long:         resourceCmd.Long,
		Example:      resourceCmd.Example,
		Aliases:      resourceCmd.Aliases,
		SilenceUsage: true,
	}

	// Create our custom RunE that calls kubectl with forced flags
	wrapperCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return executeKubectlCreate(cmd, args, kubeconfigPath, resourceType)
	}

	// Copy all flags from the resource command
	wrapperCmd.Flags().AddFlagSet(resourceCmd.Flags())

	return wrapperCmd
}

// executeKubectlCreate executes the kubectl create command with forced --dry-run=client -o yaml flags.
func executeKubectlCreate(
	cmd *cobra.Command,
	args []string,
	kubeconfigPath, resourceType string,
) error {
	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create a fresh kubectl client and command
	client := kubectl.NewClient(ioStreams)
	createCmd := client.CreateCreateCommand(kubeconfigPath)

	// Find the resource command again with the new IO streams
	freshResourceCmd := findResourceCommand(createCmd, resourceType)
	if freshResourceCmd == nil {
		return fmt.Errorf("%w: %s", ErrResourceCommandNotFound, resourceType)
	}

	// Force --dry-run=client and -o yaml FIRST before copying user flags
	err := freshResourceCmd.Flags().Set("dry-run", "client")
	if err != nil {
		return fmt.Errorf("failed to set dry-run flag: %w", err)
	}

	err = freshResourceCmd.Flags().Set("output", "yaml")
	if err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}

	// Copy all flags from wrapper to resource command (these will override defaults but not our forced flags above)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip the flags we want to force
		if flag.Name == "dry-run" || flag.Name == "output" {
			return
		}
		// Set the flag on the resource command
		if freshResourceCmd.Flags().Lookup(flag.Name) != nil {
			_ = freshResourceCmd.Flags().Set(flag.Name, flag.Value.String())
		}
	})

	// Call the fresh kubectl command's RunE or Run
	if freshResourceCmd.RunE != nil {
		err = freshResourceCmd.RunE(freshResourceCmd, args)
		if err != nil {
			return fmt.Errorf("kubectl command execution failed: %w", err)
		}

		return nil
	}

	if freshResourceCmd.Run != nil {
		freshResourceCmd.Run(freshResourceCmd, args)

		return nil
	}

	return fmt.Errorf("%w: %s", ErrNoRunFunction, resourceType)
}

// findResourceCommand finds a kubectl create subcommand by resource type name.
func findResourceCommand(createCmd *cobra.Command, resourceType string) *cobra.Command {
	for _, subCmd := range createCmd.Commands() {
		if subCmd.Name() == resourceType {
			return subCmd
		}
	}

	return nil
}
