# pkg/io/generator/yaml

This package provides generic YAML resource generators for KSail.

## Purpose

Generates generic YAML configuration files and Kubernetes resources. This package provides utilities for creating YAML-formatted configurations that are not specific to any particular Kubernetes distribution.

## Features

- **Generic YAML Generation**: Creates YAML files for various Kubernetes resources
- **Template Support**: Supports templating for dynamic YAML generation
- **Format Validation**: Ensures generated YAML is properly formatted and valid
- **Multi-Document Support**: Handles YAML files with multiple documents

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"

// Generate generic YAML resources
generator := yaml.NewGenerator(/* configuration */)
yamlContent, err := generator.Generate(/* parameters */)
if err != nil {
    log.Fatal("Failed to generate YAML:", err)
}
```

This generator is used when KSail needs to create generic YAML configurations that are not tied to a specific Kubernetes distribution or platform.

---

[⬅️ Go Back](../README.md)