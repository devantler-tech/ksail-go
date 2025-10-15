package commandrunner

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCobraCommandRunner_RunPropagatesStdout(t *testing.T) {
	t.Parallel()

	runner := NewCobraCommandRunner()
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("hello world")
		},
	}

	res, err := runner.Run(context.Background(), cmd, nil)
	require.NoError(t, err)
	require.Contains(t, res.Stdout, "hello world")
}

func TestMergeCommandError_AppendsStdStreams(t *testing.T) {
	base := errors.New("base error")
	res := CommandResult{Stdout: "info", Stderr: "fail"}

	err := MergeCommandError(base, res)
	require.ErrorContains(t, err, "base error")
	require.ErrorContains(t, err, "info")
	require.ErrorContains(t, err, "fail")
}
