# pkg/io/marshaller

This package provides data marshalling utilities for KSail.

## Purpose

Provides utilities for marshalling and unmarshalling data to and from various formats. This package abstracts the serialization/deserialization logic for different data formats used throughout KSail.

## Features

- **Format Abstraction**: Provides a common interface for different marshalling formats
- **Type Safety**: Ensures type-safe marshalling and unmarshalling operations
- **Error Handling**: Proper error reporting for marshalling operations
- **Extensible**: Easy to add support for new data formats

## Subpackages

- **[pkg/io/marshaller/yaml/](./yaml/README.md)** - YAML marshalling and unmarshalling utilities

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/marshaller"

// Use specific marshaller implementations
// See individual subpackage documentation for detailed usage
```

The marshaller package provides a consistent interface for data serialization across different formats, enabling KSail to work with various configuration and data formats seamlessly.

---

[⬅️ Go Back](../README.md)