package testutils //nolint:testpackage // Access internal helpers for focused coverage tests.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

var (
	errEmptyConfig  = errors.New("empty config content")
	errInvalidState = errors.New("invalid config content")
)

// sampleConfig models a minimal configuration structure for testing helpers.
type sampleConfig struct {
	Name string
}

// fakeConfigManager is a lightweight ConfigManager implementation that
// reads configuration content from disk and supports cached reloads. It lets
// the helper tests exercise success, error, and caching paths without relying
// on the full production managers.
type fakeConfigManager struct {
	path   string
	config *sampleConfig
}

func newFakeConfigManager(path string) configmanager.ConfigManager[sampleConfig] {
	return &fakeConfigManager{path: path}
}

func (m *fakeConfigManager) LoadConfig(_ timer.Timer) error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	content := strings.TrimSpace(string(data))
	switch {
	case content == "":
		return errEmptyConfig
	case strings.Contains(content, "invalid"):
		if m.config != nil {
			// Simulate cached response when the on-disk content becomes invalid.
			return nil
		}

		return errInvalidState
	default:
		m.config = &sampleConfig{Name: content}
	}

	return nil
}

func (m *fakeConfigManager) GetConfig() *sampleConfig {
	return m.config
}

func TestRunConfigManagerTests(t *testing.T) {
	t.Parallel()

	var validationCalled atomic.Bool

	t.Cleanup(func() {
		if !validationCalled.Load() {
			t.Fatalf("expected validation function to be invoked")
		}
	})

	scenarios := []TestScenario[sampleConfig]{
		{
			Name:                "valid config loads",
			ConfigContent:       "alpha",
			UseCustomConfigPath: true,
			ValidationFunc: func(t *testing.T, cfg *sampleConfig) {
				t.Helper()

				if cfg == nil {
					t.Fatal("expected configuration to be loaded")
				}

				if cfg.Name != "alpha" {
					t.Fatalf("expected name alpha, got %s", cfg.Name)
				}

				validationCalled.Store(true)
			},
		},
		{
			Name:        "missing file triggers error",
			ShouldError: true,
		},
	}

	RunConfigManagerTests(t, newFakeConfigManager, scenarios)
}

func TestAssertConfigManagerCaches(t *testing.T) {
	t.Parallel()

	AssertConfigManagerCaches(t, "config.yaml", "alpha", newFakeConfigManager)
}

func TestSetupTestConfigPath(t *testing.T) {
	t.Parallel()

	t.Run("returns setup func path", func(t *testing.T) {
		t.Parallel()

		expected := filepath.Join(t.TempDir(), "from-setup.yaml")
		scenario := TestScenario[sampleConfig]{
			Name: "custom setup",
			SetupFunc: func(*testing.T) string {
				return expected
			},
		}

		path := setupTestConfigPath(t, scenario)
		if path != expected {
			t.Fatalf("expected setup path %s, got %s", expected, path)
		}
	})

	t.Run("creates custom config file with content", func(t *testing.T) {
		t.Parallel()

		scenario := TestScenario[sampleConfig]{
			Name:                "custom file",
			UseCustomConfigPath: true,
			ConfigContent:       "alpha",
		}

		path := setupTestConfigPath(t, scenario)

		data, err := os.ReadFile(path) //nolint:gosec // Test controls temp file path.
		if err != nil {
			t.Fatalf("expected config file to exist: %v", err)
		}

		if strings.TrimSpace(string(data)) != "alpha" {
			t.Fatalf("unexpected file content: %q", data)
		}
	})

	t.Run("defaults to non-existent path", func(t *testing.T) {
		t.Parallel()

		scenario := TestScenario[sampleConfig]{Name: "default path"}
		path := setupTestConfigPath(t, scenario)

		if path != "non-existent-config.yaml" {
			t.Fatalf("expected default path, got %s", path)
		}
	})
}
