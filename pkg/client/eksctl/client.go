// Package eksctl provides an eksctl client implementation.
package eksctl

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/weaveworks/eksctl/pkg/ctl/cmdutils"
	"github.com/weaveworks/eksctl/pkg/ctl/create"
	"github.com/weaveworks/eksctl/pkg/ctl/delete"
	"github.com/weaveworks/eksctl/pkg/ctl/get"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

const clusterCommandName = "cluster"

// Client wraps eksctl command functionality.
type Client struct {
	ioStreams genericiooptions.IOStreams
}

// NewClient creates a new eksctl client instance.
func NewClient(ioStreams genericiooptions.IOStreams) *Client {
	return &Client{
		ioStreams: ioStreams,
	}
}

// CreateClusterCommand creates an eksctl create cluster command with all its flags and behavior.
func (c *Client) CreateClusterCommand(configFile string) *cobra.Command {
	// Create flag grouping for eksctl commands
	flagGrouping := cmdutils.NewGrouping()

	// Create the create command group
	createCmd := create.Command(flagGrouping)

	// Find the cluster subcommand
	var clusterCmd *cobra.Command

	for _, cmd := range createCmd.Commands() {
		if cmd.Name() == clusterCommandName {
			clusterCmd = cmd

			break
		}
	}

	if clusterCmd != nil {
		// Set IO streams
		clusterCmd.SetOut(c.ioStreams.Out)
		clusterCmd.SetErr(c.ioStreams.ErrOut)
		clusterCmd.SetIn(c.ioStreams.In)

		// If config file is specified, add it to the args
		if configFile != "" {
			_ = clusterCmd.Flags().Set("config-file", configFile)
		}
	}

	return clusterCmd
}

// DeleteClusterCommand creates an eksctl delete cluster command with all its flags and behavior.
func (c *Client) DeleteClusterCommand(clusterName string) *cobra.Command {
	// Create flag grouping for eksctl commands
	flagGrouping := cmdutils.NewGrouping()

	// Create the delete command group
	deleteCmd := delete.Command(flagGrouping)

	// Find the cluster subcommand
	var clusterCmd *cobra.Command

	for _, cmd := range deleteCmd.Commands() {
		if cmd.Name() == clusterCommandName {
			clusterCmd = cmd

			break
		}
	}

	if clusterCmd != nil {
		// Set IO streams
		clusterCmd.SetOut(c.ioStreams.Out)
		clusterCmd.SetErr(c.ioStreams.ErrOut)
		clusterCmd.SetIn(c.ioStreams.In)

		// Set cluster name if provided
		if clusterName != "" {
			_ = clusterCmd.Flags().Set("name", clusterName)
		}
	}

	return clusterCmd
}

// GetClusterCommand creates an eksctl get cluster command with all its flags and behavior.
func (c *Client) GetClusterCommand() *cobra.Command {
	// Create flag grouping for eksctl commands
	flagGrouping := cmdutils.NewGrouping()

	// Create the get command group
	getCmd := get.Command(flagGrouping)

	// Find the cluster subcommand
	var clusterCmd *cobra.Command

	for _, cmd := range getCmd.Commands() {
		if cmd.Name() == clusterCommandName {
			clusterCmd = cmd

			break
		}
	}

	if clusterCmd != nil {
		// Set IO streams
		clusterCmd.SetOut(c.ioStreams.Out)
		clusterCmd.SetErr(c.ioStreams.ErrOut)
		clusterCmd.SetIn(c.ioStreams.In)
	}

	return clusterCmd
}

// ExecuteClusterCommand executes an eksctl cluster command and returns the output.
func (c *Client) ExecuteClusterCommand(cmd *cobra.Command, args []string) (string, error) {
	// Capture output
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return errBuf.String(), err
	}

	return outBuf.String(), nil
}

// ListClusters lists all EKS clusters using eksctl get cluster command.
func (c *Client) ListClusters() ([]string, error) {
	// Create a buffer to capture output
	var outBuf bytes.Buffer

	tempStreams := genericiooptions.IOStreams{
		In:     c.ioStreams.In,
		Out:    &outBuf,
		ErrOut: io.Discard,
	}

	tempClient := &Client{ioStreams: tempStreams}
	cmd := tempClient.GetClusterCommand()

	// Execute the command
	if cmd == nil {
		return []string{}, nil
	}

	err := cmd.Execute()
	if err != nil {
		// If there's an error, return it
		return nil, fmt.Errorf("failed to execute eksctl get cluster: %w", err)
	}

	// Parse output to extract cluster names
	output := outBuf.String()
	clusters := parseClusterNames(output)

	return clusters, nil
}

// parseClusterNames parses the output of eksctl get cluster to extract cluster names.
func parseClusterNames(output string) []string {
	// Simple parsing - assumes output has one cluster per line
	// This is a basic implementation
	var clusters []string

	lines := bytes.Split([]byte(output), []byte("\n"))

	for i, line := range lines {
		// Skip header line
		if i == 0 {
			continue
		}

		lineStr := string(bytes.TrimSpace(line))
		if lineStr == "" {
			continue
		}

		// Extract first column (cluster name)
		fields := bytes.Fields(line)
		if len(fields) > 0 {
			clusters = append(clusters, string(fields[0]))
		}
	}

	return clusters
}
