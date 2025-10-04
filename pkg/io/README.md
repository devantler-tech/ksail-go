# pkg/io

This package provides utilities for input and output operations in KSail.

## Purpose

Provides safe and secure file I/O operations with built-in protection against path traversal attacks and other security vulnerabilities. The package includes utilities for reading and writing files safely within specified base directories, along with helpers for working with filesystem paths.

## Features

- **Safe File Reading**: `ReadFileSafe` prevents path traversal attacks by ensuring files are within the specified base directory
- **Path Validation**: Resolves and validates file paths to prevent access outside intended directories
- **Home Directory Expansion**: `ExpandHomePath` converts `~/` prefixes into the user's absolute home directory while preserving other paths unchanged
- **Security**: Protects against accidental file inclusion and malicious path manipulation
- **Clean Path Handling**: Automatically cleans and normalizes file paths

## Security

The package includes `ErrPathOutsideBase` error to indicate when a file path is outside the specified base directory, providing protection against:

- Path traversal attacks (e.g., `../../../etc/passwd`)
- Symlink attacks
- Accidental access to system files

## Subpackages

- **[pkg/io/generator/](./generator/README.md)** - Resource generation utilities
- **[pkg/io/marshaller/](./marshaller/README.md)** - Data marshalling utilities

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io"

// Safely read a file within a base directory
baseDir := "/safe/working/directory"
filePath := "config/settings.yaml"

data, err := io.ReadFileSafe(baseDir, filePath)
if err != nil {
    if errors.Is(err, io.ErrPathOutsideBase) {
        log.Fatal("Security violation: file outside base directory")
    }
    log.Fatal("Failed to read file:", err)
}
```

---

[⬅️ Go Back](../README.md)
