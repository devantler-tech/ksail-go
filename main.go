// Package main is the entry point for the KSail application.
package main

import (
	"os"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
)

//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCmd(version, commit, date)

	err := rootCmd.Execute()
	if err != nil {
		notify.Errorln(rootCmd.ErrOrStderr(), err)
		os.Exit(1)
	}

	os.Exit(0)
}

// test comment
