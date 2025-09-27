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

// run executes the main application logic and returns an exit code.
// This function is separated from main() to make it testable.
func run() int {
	rootCmd := cmd.NewRootCmd(version, commit, date)

	err := rootCmd.Execute()
	if err != nil {
		notify.Errorln(rootCmd.ErrOrStderr(), err)

		return 1
	}

	return 0
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			notify.Errorln(os.Stderr, r)
			os.Exit(1)
		}
	}()
	os.Exit(run())
}
