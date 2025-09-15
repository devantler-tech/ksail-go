# yaml

This package provides YAML marshalling and unmarshalling utilities for KSail.

## Purpose

Implements YAML-specific marshalling and unmarshalling functionality, providing utilities for converting Go data structures to and from YAML format. This is essential for working with Kubernetes manifests and configuration files.

## Features

- **YAML Marshalling**: Convert Go structs to YAML format
- **YAML Unmarshalling**: Parse YAML data into Go structs
- **Kubernetes Compatibility**: Handles YAML formats commonly used in Kubernetes
- **Multi-Document Support**: Supports YAML files with multiple documents
- **Validation**: Ensures YAML format validity

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"

// Marshal Go struct to YAML
data := MyStruct{Field: "value"}
yamlBytes, err := yaml.Marshal(data)
if err != nil {
    log.Fatal("Failed to marshal to YAML:", err)
}

// Unmarshal YAML to Go struct
var result MyStruct
err = yaml.Unmarshal(yamlBytes, &result)
if err != nil {
    log.Fatal("Failed to unmarshal YAML:", err)
}
```

This package is essential for KSail's interaction with Kubernetes resources and configuration files, which are typically represented in YAML format.

---

[⬅️ Go Back](../README.md)