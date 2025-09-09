# ASCII Art Testing - 100% Code Coverage

This directory contains comprehensive tests for the ASCII art logo functionality, achieving maximum possible code coverage through the public API.

## Current Coverage

With the standard tests (`ksail_logo_test.go`), the package achieves **92% code coverage**. The remaining 8% consists of edge cases in internal helper functions that cannot be triggered by the current embedded logo content.

## Edge Cases

The uncovered edge cases are in the internal functions:

1. **`printGreenBlueCyanPart`** - When line length < 38 characters (`cyanStartIndex`)
2. **`printGreenCyanPart`** - When line length < 32 characters (`greenCyanSplitIndex`)

These edge cases exist to handle ASCII art lines that are shorter than expected, providing robustness even though the current logo doesn't trigger them.

## Achieving 100% Coverage

Since the logo content is embedded at compile time using `//go:embed`, testing these edge cases through the public API requires modifying the embedded file and rebuilding. 

### Method 1: Automated Script

Run the provided script to automatically achieve 100% coverage:

```bash
./scripts/test_edge_cases.sh
```

This script:
1. Backs up the original logo file
2. Creates a modified logo with shorter lines that trigger the edge cases
3. Runs the tests with coverage reporting
4. Restores the original logo file

### Method 2: Manual Steps

1. Backup the original logo:
   ```bash
   cp cmd/ui/asciiart/ksail_logo.txt cmd/ui/asciiart/ksail_logo.txt.backup
   ```

2. Create a logo with shorter lines (see `scripts/test_edge_cases.sh` for example content)

3. Run tests with coverage:
   ```bash
   go test -coverprofile=coverage.out ./cmd/ui/asciiart
   go tool cover -func=coverage.out
   ```

4. Restore the original logo:
   ```bash
   mv cmd/ui/asciiart/ksail_logo.txt.backup cmd/ui/asciiart/ksail_logo.txt
   ```

## Test Structure

### `TestPrintKSailLogo`
- Basic snapshot test of the public API with original logo
- Ensures output consistency and correctness

### `TestPrintKSailLogo_Comprehensive` 
- Comprehensive testing of various aspects through public API:
  - Output validation (non-empty, contains expected elements)
  - Line structure verification
  - Consistency between calls
  - Proper formatting

### `TestPrintKSailLogo_Writers`
- Tests the function with different `io.Writer` implementations
- Ensures compatibility with various output destinations
- Tests buffer scenarios (fresh, pre-filled, large capacity)

## Design Considerations

The current approach prioritizes:

1. **Public API Testing**: All tests use only the exported `PrintKSailLogo` function
2. **No Source Modification**: Tests don't add code to the source package
3. **Comprehensive Coverage**: Maximum coverage achievable through public interface
4. **Edge Case Documentation**: Clear documentation of limitations and workarounds

## Limitations

- **Embedded Content Constraint**: The `//go:embed` directive makes runtime content injection impossible
- **Edge Case Testing**: Some internal edge cases require build-time file modification to test
- **CI/CD Considerations**: Automated edge case testing requires careful file management

This approach balances comprehensive testing with clean separation between test and source code, while providing tools to achieve 100% coverage when needed.