package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
)

type sampleConfig struct {
	Name string `yaml:"name"`
}

func defaultSampleConfig() sampleConfig {
	return sampleConfig{Name: "default"}
}

func TestLoadConfigFromFilePermissionError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("name: blocked"), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if err := os.Chmod(path, 0); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err := LoadConfigFromFile(path, defaultSampleConfig)
	if err == nil {
		t.Fatal("expected read failure")
	}

	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type stubValidator struct {
	result *validator.ValidationResult
}

func (s stubValidator) Validate(sampleConfig) *validator.ValidationResult {
	return s.result
}

func TestLoadAndValidateConfigErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("name: ok"), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	invalid := &validator.ValidationResult{
		Valid:  false,
		Errors: []validator.ValidationError{{Field: "name"}},
	}
	_, err := LoadAndValidateConfig(path, defaultSampleConfig, stubValidator{result: invalid})
	if err == nil {
		t.Fatal("expected validation failure")
	}

	if !strings.Contains(err.Error(), "failed to validate config") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFormatValidationWarningsHighlightsField(t *testing.T) {
	t.Parallel()

	result := &validator.ValidationResult{
		Warnings: []validator.ValidationError{{Field: "spec.example", Message: "check"}},
	}

	warnings := FormatValidationWarnings(result)
	if len(warnings) != 1 {
		t.Fatalf("expected single warning, got %d", len(warnings))
	}

	if !strings.Contains(warnings[0], "spec.example") {
		t.Fatalf("warning missing field context: %q", warnings[0])
	}
}
