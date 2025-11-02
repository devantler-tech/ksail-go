// Package kubectl provides a Generator implementation that wraps kubectl create commands
// with forced --dry-run=client -o yaml flags to generate Kubernetes resource manifests.
package kubectl

import (
	"errors"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
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

// Generator generates Kubernetes resource manifests using kubectl create with --dry-run=client -o yaml.
type Generator struct {
	kubeconfigPath string
	resourceType   string
}

// NewGenerator creates a new kubectl-based generator for a specific resource type.
func NewGenerator(kubeconfigPath, resourceType string) *Generator {
	return &Generator{
		kubeconfigPath: kubeconfigPath,
		resourceType:   resourceType,
	}
}

// Generate creates a gen subcommand that wraps kubectl create with forced --dry-run=client -o yaml.
func (g *Generator) Generate() *cobra.Command {
	// Create temporary kubectl client to get the resource command
	tempIOStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	tempClient := kubectl.NewClient(tempIOStreams)
	tempCreateCmd := tempClient.CreateCreateCommand(g.kubeconfigPath)

	// Find the subcommand for this resource type
	var resourceCmd *cobra.Command

	for _, subCmd := range tempCreateCmd.Commands() {
		if subCmd.Name() == g.resourceType {
			resourceCmd = subCmd

			break
		}
	}

	if resourceCmd == nil {
		panic(fmt.Sprintf("kubectl create %s command not found", g.resourceType))
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
		return g.executeKubectlCreate(cmd, args)
	}

	// Copy all flags from the resource command
	wrapperCmd.Flags().AddFlagSet(resourceCmd.Flags())

	return wrapperCmd
}

// executeKubectlCreate executes the kubectl create command with forced --dry-run=client -o yaml flags.
func (g *Generator) executeKubectlCreate(
	cmd *cobra.Command,
	args []string,
) error {
	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create a fresh kubectl client and command
	client := kubectl.NewClient(ioStreams)
	createCmd := client.CreateCreateCommand(g.kubeconfigPath)

	// Find the resource command again with the new IO streams
	freshResourceCmd := findResourceCommand(createCmd, g.resourceType)
	if freshResourceCmd == nil {
		return fmt.Errorf("%w: %s", ErrResourceCommandNotFound, g.resourceType)
	}

	// Force --dry-run=client and -o yaml FIRST before copying user flags
	err := setForcedFlags(freshResourceCmd)
	if err != nil {
		return err
	}

	// Copy all flags from wrapper to resource command
	copyUserFlags(cmd, freshResourceCmd)

	// Execute the kubectl command
	return runKubectlCommand(freshResourceCmd, args, g.resourceType)
}

// setForcedFlags sets the --dry-run=client and -o yaml flags that cannot be overridden.
func setForcedFlags(cmd *cobra.Command) error {
	err := cmd.Flags().Set("dry-run", "client")
	if err != nil {
		return fmt.Errorf("failed to set dry-run flag: %w", err)
	}

	err = cmd.Flags().Set("output", "yaml")
	if err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}

	return nil
}

// copyUserFlags copies user-provided flags from wrapper command to kubectl command.
func copyUserFlags(wrapperCmd, targetCmd *cobra.Command) {
	wrapperCmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip the flags we want to force
		if flag.Name == "dry-run" || flag.Name == "output" {
			return
		}

		// Set the flag on the resource command
		targetFlag := targetCmd.Flags().Lookup(flag.Name)
		if targetFlag != nil {
			copyFlagValue(flag, targetCmd)
		}
	})
}

// copyFlagValue copies a flag value, handling slice flags specially.
func copyFlagValue(flag *pflag.Flag, targetCmd *cobra.Command) {
	// For slice flags, we need to get the actual slice values
	if sliceVal, ok := flag.Value.(pflag.SliceValue); ok {
		strSlice := sliceVal.GetSlice()
		for _, v := range strSlice {
			_ = targetCmd.Flags().Set(flag.Name, v)
		}
	} else {
		// For non-slice flags, just copy the string value
		_ = targetCmd.Flags().Set(flag.Name, flag.Value.String())
	}
}

// runKubectlCommand executes the kubectl command's Run or RunE function.
func runKubectlCommand(cmd *cobra.Command, args []string, resourceType string) error {
	if cmd.RunE != nil {
		err := cmd.RunE(cmd, args)
		if err != nil {
			return fmt.Errorf("kubectl command execution failed: %w", err)
		}

		return nil
	}

	if cmd.Run != nil {
		cmd.Run(cmd, args)

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
