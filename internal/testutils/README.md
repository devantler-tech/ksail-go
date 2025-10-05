# internal/testutils

This package provides general-purpose testing utilities for KSail's internal packages.

## Purpose

Contains generic testing helpers, utilities, and shared functionality used across multiple packages for testing. This package provides foundational testing capabilities that support testing in various parts of KSail's codebase.

## Features

- **File Testing Utilities**: Helpers for file operations and temporary file management
- **Error Handling**: Shared test error utilities and error testing patterns
- **String Utilities**: String manipulation and testing helpers
- **Marshal/Unmarshal Testing**: Utilities for testing serialization and deserialization
- **Name Case Testing**: Utilities for testing naming conventions and case conversions
- **Generic Helpers**: Common testing patterns and setup/teardown utilities

## Usage

```go
import "github.com/devantler-tech/ksail-go/internal/testutils"

// Use file testing utilities
tempFile := testutils.CreateTempFile(t, "test-content")
defer os.Remove(tempFile)

// Use error testing utilities
testutils.AssertError(t, err, "expected error message")

// Use string testing utilities
result := testutils.NormalizeString(input)
```

## Key Components

- **file.go**: File operation testing utilities
- **errors.go**: Error handling and testing patterns
- **helpers.go**: General-purpose testing helper functions
- **marshal.go**: Serialization/deserialization testing utilities
- **namecases.go**: Name case conversion testing utilities
- **string.go**: String manipulation testing helpers

This package provides the foundational testing infrastructure used throughout KSail's test suite.

---

[⬅️ Go Back](../../README.md)
