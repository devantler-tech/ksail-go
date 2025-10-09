// Package cliwrapper provides utilities for wrapping urfave/cli applications in Cobra commands.
package cliwrapper

import (
	"bytes"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/urfave/cli" //nolint:depguard // This package specifically wraps urfave/cli apps
)

// WrapCliApp wraps a urfave/cli.App into a Cobra command.
// This allows embedding urfave/cli applications within Cobra-based CLIs.
func WrapCliApp(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:                app.Name,
		Short:              app.Usage,
		Long:               app.UsageText,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			// Prepend the app name to args to match urfave/cli expectations
			cliArgs := append([]string{app.Name}, args...)

			return app.Run(cliArgs)
		},
		SilenceUsage: true,
	}

	return cmd
}

// WrapCliAppWithIO wraps a urfave/cli.App into a Cobra command with custom IO streams.
// This is useful when you need to control stdin/stdout/stderr.
func WrapCliAppWithIO(app *cli.App, stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:                app.Name,
		Short:              app.Usage,
		Long:               app.UsageText,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			// Save original streams
			origStdin, origStdout, origStderr := os.Stdin, os.Stdout, os.Stderr

			// Create pipes for IO redirection
			if stdin != nil {
				if r, ok := stdin.(*os.File); ok {
					os.Stdin = r
				}
			}

			if stdout != nil {
				if w, ok := stdout.(*os.File); ok {
					os.Stdout = w
				} else {
					// For non-file writers, we need to handle differently
					app.Writer = stdout
				}
			}

			if stderr != nil {
				if w, ok := stderr.(*os.File); ok {
					os.Stderr = w
				} else {
					app.ErrWriter = stderr
				}
			}

			// Restore streams when done
			defer func() {
				os.Stdin, os.Stdout, os.Stderr = origStdin, origStdout, origStderr
			}()

			// Prepend the app name to args to match urfave/cli expectations
			cliArgs := append([]string{app.Name}, args...)

			return app.Run(cliArgs)
		},
		SilenceUsage: true,
	}

	return cmd
}

// CaptureCliAppOutput runs a urfave/cli.App and captures its output.
// This is useful for testing.
func CaptureCliAppOutput(app *cli.App, args []string) (string, string, error) {
	var outBuf, errBuf bytes.Buffer

	app.Writer = &outBuf
	app.ErrWriter = &errBuf

	cliArgs := append([]string{app.Name}, args...)
	err := app.Run(cliArgs)

	return outBuf.String(), errBuf.String(), err
}
