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
	exitCode := runSafely()

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func runSafely() int {
	exitCode := 0

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			panicMessage := fmt.Sprintf("panic recovered: %v\n%s", r, stack)
			notify.Errorln(os.Stderr, panicMessage)

			exitCode = 1
		}
	}()

	exitCode = run()

	return exitCode
}
