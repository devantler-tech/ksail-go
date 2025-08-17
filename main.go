// Package main is the entry point for the KSail application.
package main

import (
	"os"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	err := rootCmd.Execute()
	if err != nil {
		notify.Errorln(err)
		os.Exit(1)
	}

	os.Exit(0)
}
