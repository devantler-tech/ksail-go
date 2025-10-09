package cliwrapper_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/cliwrapper"
	"github.com/urfave/cli"
)

const (
	testAppName  = "testapp"
	testAppUsage = "A test application"
)

func TestWrapCliApp(t *testing.T) {
	t.Parallel()

	app := cli.NewApp()
	app.Name = testAppName
	app.Usage = testAppUsage

	cmd := cliwrapper.WrapCliApp(app)

	if cmd.Use != testAppName {
		t.Errorf("expected Use to be %q, got %q", testAppName, cmd.Use)
	}

	if cmd.Short != testAppUsage {
		t.Errorf("expected Short to be %q, got %q", testAppUsage, cmd.Short)
	}

	if !cmd.DisableFlagParsing {
		t.Error("expected DisableFlagParsing to be true")
	}
}

func TestWrapCliAppWithIO(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	app := cli.NewApp()
	app.Name = testAppName
	app.Usage = testAppUsage
	app.Action = func(c *cli.Context) error {
		_, _ = c.App.Writer.Write([]byte("test output"))

		return nil
	}

	cmd := cliwrapper.WrapCliAppWithIO(app, nil, &stdout, &stderr)

	if cmd.Use != testAppName {
		t.Errorf("expected Use to be %q, got %q", testAppName, cmd.Use)
	}
}

func TestCaptureCliAppOutput(t *testing.T) {
	t.Parallel()

	app := cli.NewApp()
	app.Name = testAppName
	app.Action = func(c *cli.Context) error {
		_, _ = c.App.Writer.Write([]byte("stdout output"))
		_, _ = c.App.ErrWriter.Write([]byte("stderr output"))

		return nil
	}

	stdout, stderr, err := cliwrapper.CaptureCliAppOutput(app, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout != "stdout output" {
		t.Errorf("expected stdout to be 'stdout output', got %q", stdout)
	}

	if stderr != "stderr output" {
		t.Errorf("expected stderr to be 'stderr output', got %q", stderr)
	}
}
