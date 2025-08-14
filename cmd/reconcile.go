/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// reconcileCmd represents the reconcile command.
var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleReconcile()
	},
}

// --- internals ---

func handleReconcile() error {
	if err := InitServices(); err != nil {
		return err
	}

	err := configValidator.Validate()
	if err != nil {
		return err
	}
  // TODO: Validate workloads
  // TODO: Reconcile
	return nil
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}
