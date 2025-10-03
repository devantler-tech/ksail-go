# Quickstart: CLI Command Timing

## Purpose

This quickstart guide validates that the CLI command timing feature is working correctly by testing the primary user scenarios from the feature specification.

## Prerequisites

- KSail built from source with timing feature implemented
- Docker running (for cluster commands)
- Clean test environment (no existing clusters)

## Test Scenario 1: Multi-Stage Command (cluster up)

**Objective**: Verify timing display for multi-stage commands with progressive updates.

**Steps**:

```bash
# Run a multi-stage command
./ksail cluster up

# Expected output (timing values will vary):
# Creating cluster... [stage: 2.5s|total: 2.5s]
# Installing CNI... [stage: 1.7s|total: 4.2s]
# Deploying controllers... [stage: 4.3s|total: 8.5s]
# Cluster ready [stage: 1.6s|total: 10.1s]
```

**Success Criteria**:

- ✅ Timing displayed after each stage completes
- ✅ Format matches `[stage: X|total: Y]` pattern
- ✅ Total time increases progressively
- ✅ Stage time resets for each new stage
- ✅ Final success message includes timing

**Validation**:

```bash
# Verify timing format in output
./ksail cluster up 2>&1 | grep -E '\[[0-9.]+[a-z]+ total\|[0-9.]+[a-z]+ stage\]'
# Should return multiple matches

# Clean up
./ksail cluster down
```

## Test Scenario 2: Single-Stage Command (init)

**Objective**: Verify simplified timing format for single-stage commands.

**Steps**:

```bash
# Navigate to a temporary directory
cd /tmp/ksail-test

# Run a single-stage command
./ksail init --distribution Kind

# Expected output (timing will vary):
# Initialized KSail project [1.2s]
```

**Success Criteria**:

- ✅ Timing displayed in simplified format
- ✅ Format matches `[stage: X]` pattern (no "total" split for single-stage)
- ✅ Success message includes timing
- ✅ Sub-second precision visible (e.g., "1.2s", "500ms")

**Validation**:

```bash
# Verify simplified timing format
../../../../ksail init --distribution Kind 2>&1 | grep -E '\[[0-9.]+[a-z]+\]$'
# Should return one match with simplified format

# Clean up
cd -
rm -rf /tmp/ksail-test
```

## Test Scenario 3: Command Failure (No Timing)

**Objective**: Verify timing is NOT displayed on command failures.

**Steps**:

```bash
# Run a command that will fail (invalid distribution)
./ksail init --distribution InvalidDistribution

# Expected output:
# Error: unsupported distribution "InvalidDistribution"
# (NO timing information)
```

**Success Criteria**:

- ✅ Error message displayed
- ✅ NO timing information in error output
- ✅ Command exits with non-zero code

**Validation**:

```bash
# Verify no timing in error output
./ksail init --distribution InvalidDistribution 2>&1 | grep -E '\[[0-9.]+[a-z]+.*\]'
# Should return NO matches

# Verify non-zero exit code
./ksail init --distribution InvalidDistribution
echo $?  # Should be non-zero (e.g., 1)
```

## Test Scenario 4: Long-Running Command (cluster down)

**Objective**: Verify timing handles longer durations correctly.

**Steps**:

```bash
# First create a cluster
./ksail cluster up

# Then tear it down (longer operation)
./ksail cluster down

# Expected output (timing will vary):
# Stopping cluster... [stage: 1.5s|total: 1.5s]
# Deleting resources... [stage: 1.7s|total: 3.2s]
# Cluster deleted [stage: 1.9s|total: 5.1s]
```

**Success Criteria**:

- ✅ Timing displays with appropriate units (seconds, not milliseconds for longer ops)
- ✅ Progressive timing updates during teardown
- ✅ Format consistent with multi-stage pattern

**Validation**:

```bash
# Verify timing uses seconds/minutes for longer operations
./ksail cluster down 2>&1 | grep -E '\[[0-9]+[sm].*\]'
# Should match seconds or minutes format
```

## Test Scenario 5: Very Fast Command

**Objective**: Verify sub-millisecond and millisecond precision.

**Steps**:

```bash
# Run a very fast command (help or version)
./ksail --version

# Expected output:
# ksail version 0.x.x [123µs]
# or
# ksail version 0.x.x [2ms]
```

**Success Criteria**:

- ✅ Timing displays for even very fast operations
- ✅ Precision appropriate for sub-second durations (µs or ms)
- ✅ Format remains consistent

**Validation**:

```bash
# Verify timing appears even for fast commands
./ksail --version 2>&1 | grep -E '\[[0-9.]+[µm]s\]'
# Should match microseconds or milliseconds
```

## Performance Validation

**Objective**: Verify timing mechanism adds <1ms overhead.

**Steps**:

```bash
# Run a command multiple times and compare with baseline
# (This requires instrumentation in the code for precise measurement)

# Baseline: Run without timing (would require feature flag)
time ./ksail --version

# With timing: Run with timing feature
time ./ksail --version

# Compare the results
```

**Success Criteria**:

- ✅ Timing overhead is negligible (<1ms)
- ✅ No noticeable performance degradation
- ✅ Memory footprint minimal (~100 bytes per timer)

## Integration Validation

**Objective**: Verify timing works across all command types.

**Commands to Test**:

```bash
# Cluster commands
./ksail cluster up
./ksail cluster status
./ksail cluster list
./ksail cluster down

# Workload commands (if implemented)
./ksail workload reconcile

# Initialization commands
./ksail init --distribution Kind
```

**Success Criteria**:

- ✅ All commands display timing on success
- ✅ No commands display timing on errors
- ✅ Format consistent across all commands
- ✅ No runtime errors or panics

## Troubleshooting

### Issue: No Timing Displayed

**Possible Causes**:

- Feature not fully integrated in command
- Timer not started before command execution
- Success message not including timing parameter

**Debug Steps**:

```bash
# Check if timer package exists
ls -la pkg/ui/timer/

# Verify notify package has FormatTiming function
grep "FormatTiming" cmd/ui/notify/notify.go

# Run with verbose output (if available)
./ksail cluster up --verbose
```

### Issue: Incorrect Timing Format

**Possible Causes**:

- FormatTiming logic incorrect
- isMultiStage flag not set correctly
- Duration.String() not used

**Debug Steps**:

```bash
# Run unit tests for timer and notify packages
go test ./pkg/ui/timer/... -v
go test ./cmd/ui/notify/... -v
```

### Issue: Timing on Errors

**Possible Causes**:

- Error path incorrectly calling Success with timing
- Timer not isolated to success paths

**Debug Steps**:

```bash
# Check error handling in commands
grep -A 5 "notify.Error" cmd/cluster/*.go
```

## Quickstart Validation Checklist

After implementation, run through all scenarios and check:

- [ ] Test Scenario 1: Multi-stage timing works ✅
- [ ] Test Scenario 2: Single-stage simplified format works ✅
- [ ] Test Scenario 3: No timing on errors ✅
- [ ] Test Scenario 4: Long-duration formatting correct ✅
- [ ] Test Scenario 5: Sub-second precision works ✅
- [ ] Performance: <1ms overhead validated ✅
- [ ] Integration: All commands tested ✅

## Success Criteria

Quickstart is considered successful when:

1. ✅ All 5 test scenarios pass
2. ✅ Performance validation confirms <1ms overhead
3. ✅ Integration validation shows timing on all commands
4. ✅ No regressions in existing command functionality
5. ✅ Timing format matches specification exactly

**Estimated Execution Time**: 10-15 minutes for complete quickstart validation.

## Next Steps

After successful quickstart validation:

1. Run full test suite: `go test ./...`
2. Run linter: `golangci-lint run`
3. Test on target platforms (Linux amd64/arm64, macOS amd64/arm64)
4. Create pull request with feature implementation
5. Update user documentation with timing feature examples
