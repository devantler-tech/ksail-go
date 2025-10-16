package commandrunner_test

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCobraCommandRunner_RunPropagatesStdout(t *testing.T) {
	t.Parallel()

	runner := commandrunner.NewCobraCommandRunner()
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("hello world")
		},
	}

	res, err := runner.Run(context.Background(), cmd, nil)
	require.NoError(t, err)
	require.Contains(t, res.Stdout, "hello world")
}

func TestMergeCommandError_AppendsStdStreams(t *testing.T) {
	t.Parallel()

	res := commandrunner.CommandResult{Stdout: "info", Stderr: "fail"}

	err := commandrunner.MergeCommandError(assert.AnError, res)
	require.ErrorContains(t, err, "base error")
	require.ErrorContains(t, err, "info")
	require.ErrorContains(t, err, "fail")
}
