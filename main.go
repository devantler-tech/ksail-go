// Package main is the entry point for the KSail application.
package main

import (
	"fmt"
	"os"
	"runtime/debug"

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
	exitCode := runSafelyWithArgs(os.Args[1:])

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func runSafelyWithArgs(args []string) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("panic recovered: %v\n%s", r, debug.Stack())
			notify.Errorln(os.Stderr, panicMessage)
			exitCode = 1
		}
	}()

	exitCode = runWithArgs(args)
	return exitCode
}

// run executes the main application logic and returns an exit code.
// This function is separated from main() to make it testable.
func run() int {
	return runWithArgs(os.Args[1:])
}

func runWithArgs(args []string) int {
	rootCmd := cmd.NewRootCmd(version, commit, date)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	if err != nil {
		notify.Errorln(rootCmd.ErrOrStderr(), err)

		return 1
	}

	return 0
}
