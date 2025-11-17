# Test Guidelines for rekap

## Overview

The rekap test suite is designed to be **fast**, **reliable**, and **comprehensive**. All tests should complete in under 1 second for the entire suite.

## Current Performance

- **Total test time**: ~0.05s (without cache), ~0.18s (with overhead)
- **Target**: < 1s for full suite
- **Status**: ✅ Exceeding target

## Test Categories

### Unit Tests
- Located in `*_test.go` files alongside the code they test
- Should test individual functions and methods in isolation
- Should use mocks/fakes for external dependencies
- Should complete in < 10ms per test

### Integration Tests
- Currently minimal (most tests are unit tests)
- Test interactions between components
- May require actual system calls (marked with t.Skip when unavailable)

## Writing Fast Tests

### 1. Enable Parallelization
All tests that don't share global state should be parallelized:

```go
func TestSomething(t *testing.T) {
    t.Parallel()  // Add this as the first line
    // ... rest of test
}
```

### 2. Use Short Timeouts
For tests that use contexts, keep timeouts minimal:

```go
// Good: 500ms is plenty for most operations
ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()

// Bad: Don't use multi-second timeouts unless absolutely necessary
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

### 3. Mock Expensive Operations
Don't make actual system calls or file I/O when possible:

```go
// Good: Use test data structures
result := CalculateFragmentation(ctx, mockApps, mockBrowsers, mockUptime, thresholds)

// Bad: Making real system calls in unit tests
result := CollectApps(ctx)  // Only OK if you check availability and skip
```

### 4. Use Table-Driven Tests
For multiple test cases, use table-driven tests:

```go
func TestFormatBytes(t *testing.T) {
    t.Parallel()
    tests := []struct {
        bytes    int64
        expected string
    }{
        {500, "500 B"},
        {1024, "1.0 KB"},
        {1048576, "1.0 MB"},
    }

    for _, tt := range tests {
        result := FormatBytes(tt.bytes)
        if result != tt.expected {
            t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
        }
    }
}
```

### 5. Skip Tests Gracefully
For tests that require specific system permissions or hardware:

```go
func TestCollectBattery(t *testing.T) {
    t.Parallel()
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    result := CollectBattery(ctx)

    if !result.Available {
        t.Skip("Battery not available (running on desktop?)")
    }

    // ... continue with assertions
}
```

## Running Tests

### Quick Test (cached)
```bash
make test-fast
# or
go test ./...
```

### Full Test (no cache)
```bash
make test
# or
go test -count=1 ./...
```

### With Coverage
```bash
make test-coverage
# Opens coverage.html in browser
```

### Run Specific Package
```bash
go test ./internal/collectors
```

### Run Specific Test
```bash
go test -run TestFormatBytes ./internal/collectors
```

### Verbose Output
```bash
go test -v ./...
```

## Test Coverage Goals

- **Overall**: > 70%
- **Collectors**: > 70% (currently 32.7%)
- **Config**: > 80% (currently 83.0% ✅)
- **Theme**: > 80% (currently 81.8% ✅)
- **UI**: > 70% (currently 32.8%)
- **Permissions**: > 50% (currently 27.9%)

## Common Patterns

### Testing Collectors
```go
func TestCollectSomething(t *testing.T) {
    t.Parallel()
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    result := CollectSomething(ctx)

    // Best-effort: skip if not available
    if !result.Available {
        t.Log("Something not available")
        return
    }

    // Assert valid values
    if result.Value < 0 {
        t.Errorf("Value should be >= 0, got %d", result.Value)
    }
}
```

### Testing Pure Functions
```go
func TestPureFunction(t *testing.T) {
    t.Parallel()
    result := PureFunction(input)
    
    if result != expected {
        t.Errorf("PureFunction(%v) = %v, want %v", input, result, expected)
    }
}
```

### Testing with Subtests
```go
func TestComplexBehavior(t *testing.T) {
    t.Parallel()
    
    t.Run("case 1", func(t *testing.T) {
        t.Parallel()
        // test code
    })
    
    t.Run("case 2", func(t *testing.T) {
        t.Parallel()
        // test code
    })
}
```

## Benchmarking

For performance-critical code, add benchmarks:

```go
func BenchmarkCalculateFragmentation(b *testing.B) {
    ctx := context.Background()
    apps := mockAppsData()
    browsers := mockBrowsersData()
    uptime := mockUptimeData()
    thresholds := DefaultFragmentationThresholds()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        CalculateFragmentation(ctx, apps, browsers, uptime, thresholds)
    }
}
```

Run benchmarks:
```bash
make test-bench
```

## CI/CD Integration

Tests run in CI on every push. They must:
- Complete in < 5 seconds
- Pass on macOS (primary target)
- Handle missing permissions gracefully
- Not require network access
- Not depend on specific hardware (battery, etc.)

## Troubleshooting

### Tests are slow
- Check if `t.Parallel()` is missing
- Look for unnecessarily long timeouts
- Check if tests are making actual system calls

### Tests are flaky
- Check for race conditions (run `go test -race`)
- Ensure tests clean up after themselves
- Don't depend on timing assumptions
- Use proper synchronization primitives

### Coverage is low
- Add tests for error paths
- Test edge cases
- Test with various input combinations
- Mock external dependencies

## Linting

Run linter before committing:
```bash
make lint
```

The golangci-lint configuration is in `.golangci.yml`. It runs multiple linters including:
- errcheck: Unchecked errors
- gosimple: Code simplification
- govet: Standard Go checks
- staticcheck: Advanced static analysis
- gosec: Security issues
- gofmt/goimports: Code formatting

## Best Practices Summary

1. ✅ Add `t.Parallel()` to all tests
2. ✅ Use short timeouts (500ms max for unit tests)
3. ✅ Skip gracefully when resources unavailable
4. ✅ Use table-driven tests for multiple cases
5. ✅ Mock expensive operations
6. ✅ Test error paths and edge cases
7. ✅ Keep tests focused and simple
8. ✅ Run tests frequently during development
9. ✅ Maintain > 70% coverage
10. ✅ Fix linting issues immediately

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Go Test Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [golangci-lint Docs](https://golangci-lint.run/)
