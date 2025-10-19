# Options for Console Logging in Kind Provisioner

## Background

The k3d provisioner shows console logs to users because it executes Cobra commands through the `commandrunner` package, which forwards stdout/stderr to the terminal using `io.MultiWriter`. The kind provisioner currently uses the Provider interface directly, which doesn't provide the same console output visibility.

## Options

### Option 1: Use Kind's Cobra Commands (with Console Logging) ⭐ RECOMMENDED

**Implementation**: Use the POC implementation with modifications to display output in real-time.

**How it works**:
- Modify `SimpleKindRunner` to use `io.MultiWriter` like k3d's command-runner
- Forward stdout/stderr to `os.Stdout`/`os.Stderr` while also capturing
- Users see real-time progress from kind's Cobra commands

**Pros**:
- ✅ Real-time console output (kind's native messages)
- ✅ Consistent UX with k3d provisioner
- ✅ Users see exactly what kind CLI shows
- ✅ No need to reimplement logging

**Cons**:
- ❌ Requires temp file handling for Create operation
- ❌ Hybrid architecture (Cobra + Docker for start/stop)
- ❌ More complex than current Provider approach

**Code changes required**:
```go
// In SimpleKindRunner.Run()
func (r *SimpleKindRunner) Run(ctx context.Context, cmd *cobra.Command, args []string) (string, string, error) {
    var outBuf, errBuf bytes.Buffer
    
    // Forward to console AND capture
    cmd.SetOut(io.MultiWriter(&outBuf, os.Stdout))
    cmd.SetErr(io.MultiWriter(&errBuf, os.Stderr))
    
    cmd.SetContext(ctx)
    cmd.SetArgs(args)
    cmd.SilenceUsage = true
    cmd.SilenceErrors = true
    
    execErr := cmd.ExecuteContext(ctx)
    return outBuf.String(), errBuf.String(), execErr
}
```

**Adoption path**:
1. Update POC implementation with console forwarding
2. Replace production `KindClusterProvisioner` with command-based version
3. Update factory to use command-based provisioner
4. Test with real clusters to verify output

---

### Option 2: Use Kind's Provider with Custom Logger

**Implementation**: Pass a custom logger to kind's Provider that forwards to ksail's notify system.

**How it works**:
- Implement kind's `log.Logger` interface
- Forward log messages to ksail's `notify` package
- Pass custom logger to Provider via `cluster.ProviderWithLogger(logger)`

**Pros**:
- ✅ Keeps current Provider-based architecture
- ✅ No temp file handling
- ✅ Consistent with kind's design intent

**Cons**:
- ❌ Requires implementing kind's Logger interface
- ❌ Log messages may not match kind CLI exactly
- ❌ More work to format messages nicely

**Code changes required**:
```go
// Implement kind's Logger interface
type KsailKindLogger struct {
    out io.Writer
}

func (l *KsailKindLogger) V(level log.Level) log.InfoLogger {
    return &ksailInfoLogger{out: l.out, level: level}
}

func (l *KsailKindLogger) Warn(message string) {
    notify.WriteMessage(notify.Message{
        Type: notify.WarningType,
        Content: message,
    })
}

// ... implement other methods

// In Create()
logger := &KsailKindLogger{out: cmd.OutOrStdout()}
provider := cluster.NewProvider(
    cluster.ProviderWithLogger(logger),
)
```

**Adoption path**:
1. Implement `log.Logger` and `log.InfoLogger` interfaces
2. Update `NewKindClusterProvisioner` to accept io.Writer for output
3. Pass custom logger to Provider in Create/Delete operations
4. Test and adjust message formatting

---

### Option 3: Wrap Provider Calls with Custom Progress Messages

**Implementation**: Keep Provider interface but add ksail's own progress messages around operations.

**How it works**:
- Use ksail's `notify` package to show progress
- Display messages like "Creating cluster...", "Pulling images...", etc.
- Kind's actual output is silenced (as it is now)

**Pros**:
- ✅ Keeps current simple architecture
- ✅ No temp files or complex changes
- ✅ Full control over UX messaging
- ✅ Minimal code changes

**Cons**:
- ❌ Doesn't show kind's detailed progress
- ❌ Messages are static, not real-time from kind
- ❌ Different UX than k3d provisioner

**Code changes required**:
```go
func (k *KindClusterProvisioner) Create(ctx context.Context, name string) error {
    target := setName(name, k.kindConfig.Name)
    
    // Show progress messages
    notify.WriteMessage(notify.Message{
        Type: notify.ActivityType,
        Content: "Preparing cluster configuration...",
    })
    
    // Show image pull progress
    notify.WriteMessage(notify.Message{
        Type: notify.ActivityType, 
        Content: "Pulling node image (this may take a few minutes)...",
    })
    
    err := k.provider.Create(
        target,
        cluster.CreateWithV1Alpha4Config(k.kindConfig),
        cluster.CreateWithDisplayUsage(true),
        cluster.CreateWithDisplaySalutation(true),
    )
    if err != nil {
        return fmt.Errorf("failed to create kind cluster: %w", err)
    }
    
    notify.WriteMessage(notify.Message{
        Type: notify.ActivityType,
        Content: "Configuring cluster networking...",
    })
    
    return nil
}
```

**Adoption path**:
1. Add notify.WriteMessage calls around key operations
2. Test to ensure messages align with actual operations
3. Adjust timing and content based on user feedback

---

### Option 4: Hybrid - Provider with Command Output Capture

**Implementation**: Use Provider interface but capture its stdout/stderr.

**How it works**:
- Redirect `os.Stdout` and `os.Stderr` temporarily
- Use `io.MultiWriter` to display AND capture
- Restore original streams after operation

**Pros**:
- ✅ Shows kind's actual console output
- ✅ Keeps Provider-based architecture
- ✅ No Cobra command complexity

**Cons**:
- ❌ Global stdout/stderr manipulation (risky)
- ❌ Need mutex to prevent concurrent operations
- ❌ Complex restoration logic

**Code changes required**:
```go
func (k *KindClusterProvisioner) Create(ctx context.Context, name string) error {
    target := setName(name, k.kindConfig.Name)
    
    // Capture and display output
    originalStdout := os.Stdout
    originalStderr := os.Stderr
    
    r, w, _ := os.Pipe()
    os.Stdout = w
    os.Stderr = w
    
    // Forward in goroutine
    go func() {
        io.Copy(io.MultiWriter(originalStdout, &captureBuffer), r)
    }()
    
    err := k.provider.Create(
        target,
        cluster.CreateWithV1Alpha4Config(k.kindConfig),
        cluster.CreateWithDisplayUsage(true),
        cluster.CreateWithDisplaySalutation(true),
    )
    
    w.Close()
    os.Stdout = originalStdout
    os.Stderr = originalStderr
    
    if err != nil {
        return fmt.Errorf("failed to create kind cluster: %w", err)
    }
    
    return nil
}
```

**Adoption path**:
1. Implement output capture wrapper
2. Add mutex for thread safety
3. Test thoroughly for edge cases
4. Monitor for issues with concurrent operations

---

## Comparison Matrix

| Option | Console Output | Complexity | Consistency | Maintenance |
|--------|---------------|------------|-------------|-------------|
| **Option 1: Cobra Commands** | ✅ Real-time | 🟡 Medium | ✅ High | 🟡 Medium |
| **Option 2: Custom Logger** | ✅ Custom | 🟡 Medium | 🟢 High | 🟢 Low |
| **Option 3: Static Messages** | 🟡 Static | 🟢 Low | 🟡 Medium | 🟢 Low |
| **Option 4: Output Capture** | ✅ Real-time | 🔴 High | 🟢 High | 🔴 High |

## Recommendation

**For best user experience**: **Option 1 (Cobra Commands)** 
- Provides the same console logging UX as k3d
- Users see real-time progress from kind itself
- Worth the additional complexity for consistent UX

**For quickest implementation**: **Option 3 (Static Messages)**
- Minimal code changes
- Low risk
- Good enough for most use cases

**For architectural purity**: **Option 2 (Custom Logger)**
- Stays aligned with kind's design
- No temp files or Cobra complexity
- Requires more implementation effort

## Implementation Priority

If adopting Option 1:
1. ✅ POC already exists (created in this PR)
2. Add console output forwarding to `SimpleKindRunner`
3. Add integration tests with real clusters
4. Update factory to use command-based provisioner
5. Document limitations (temp files, hybrid architecture)

If adopting Option 3:
1. Add notify messages to Create/Delete/Start/Stop operations
2. Test timing of messages
3. Consider adding spinner/progress bar
4. Document that output is ksail-generated, not from kind

## Next Steps

Choose an option based on:
- **UX priority**: How important is real-time kind output?
- **Maintenance commitment**: Can we maintain complex Cobra integration?
- **Timeline**: How quickly do we need this feature?

The POC implementation in this PR makes Option 1 ready to adopt with minor modifications.
