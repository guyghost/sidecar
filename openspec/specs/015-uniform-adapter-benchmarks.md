# OpenSpec: Uniform Adapter Benchmarks

**ID**: `td-64e7a8`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P3
**Type**: Chore / Testing
**Effort**: 3 story points

## Problem Statement

Only Claude Code and Codex adapters have benchmark tests (`adapter_bench_test.go`). The other 6 adapters (Cursor, Warp, Kiro, Gemini CLI, Amp, OpenCode) have no performance benchmarks. This makes it impossible to:
- Detect performance regressions in adapters
- Compare adapter efficiency
- Validate optimization efforts (e.g., `td-0da8fe` incremental parsing)

## Objective

Create a standardized benchmark framework for all 8 adapters covering session listing, message parsing, search, and large session handling.

## Constraints

- Benchmarks must be reproducible (use generated test fixtures, not live AI tool data)
- Benchmarks must not require actual AI tool installation
- Must run in CI without flakiness
- Must provide meaningful comparison across adapters

## Technical Design

### Benchmark Framework

```go
// internal/adapter/testutil/benchmark.go

// BenchmarkFixture generates test data for adapter benchmarks.
type BenchmarkFixture struct {
    SmallSession  string // 10 messages, ~10KB
    MediumSession string // 100 messages, ~500KB
    LargeSession  string // 1000 messages, ~10MB
    HugeSession   string // 10000 messages, ~100MB
}

// GenerateFixture creates a fixture directory for a specific adapter format.
func GenerateFixture(adapter string, size FixtureSize) (string, func()) {
    // Creates temp directory with realistic adapter-specific data
    // Returns path and cleanup function
}
```

### Standard Benchmark Suite

Each adapter implements the same benchmark functions:

```go
// internal/adapter/{name}/adapter_bench_test.go

func BenchmarkSessions_Small(b *testing.B)   // 5 sessions
func BenchmarkSessions_Medium(b *testing.B)  // 50 sessions
func BenchmarkSessions_Large(b *testing.B)   // 500 sessions

func BenchmarkMessages_Small(b *testing.B)   // 10 messages
func BenchmarkMessages_Medium(b *testing.B)  // 100 messages
func BenchmarkMessages_Large(b *testing.B)   // 1000 messages
func BenchmarkMessages_Huge(b *testing.B)    // 10000 messages

func BenchmarkSearch_Small(b *testing.B)     // Search across 10 messages
func BenchmarkSearch_Large(b *testing.B)     // Search across 1000 messages

func BenchmarkDetect(b *testing.B)           // Detect() latency

// Memory allocation tracking
func BenchmarkMessages_Allocs(b *testing.B)  // Track allocations per message
```

### Fixture Generation per Adapter

| Adapter | Data Format | Fixture Generator |
|---------|-------------|-------------------|
| Claude Code | JSONL files in project dirs | Generate JSONL with realistic message structure |
| Codex | JSONL in date-organized dirs | Generate YYYY/MM/DD directory structure |
| Cursor | SQLite with hex-encoded JSON | Generate populated SQLite database |
| Warp | SQLite with ai_queries table | Generate populated SQLite database |
| Kiro | SQLite with conversations_v2 | Generate populated SQLite database |
| Gemini CLI | JSON session files | Generate session-*.json files |
| Amp | JSON thread files | Generate T-{uuid}.json files |
| OpenCode | JSON files per message | Generate message directory structure |

### Benchmark Report

```
$ go test -bench=. -benchmem ./internal/adapter/...

BenchmarkSessions_Medium/claudecode-8     5000    234000 ns/op    45678 B/op   123 allocs/op
BenchmarkSessions_Medium/codex-8          3000    456000 ns/op    67890 B/op   234 allocs/op
BenchmarkSessions_Medium/cursor-8         2000    678000 ns/op    89012 B/op   345 allocs/op
...

BenchmarkMessages_Large/claudecode-8       500   2340000 ns/op   456789 B/op  1234 allocs/op
BenchmarkMessages_Large/codex-8            300   4560000 ns/op   678901 B/op  2345 allocs/op
...
```

### CI Integration

Add benchmark comparison to CI:

```yaml
# .github/workflows/go-ci.yml (new job)
benchmark:
  runs-on: ubuntu-latest
  steps:
    - run: go test -bench=. -benchmem -count=3 ./internal/adapter/... | tee bench.txt
    - uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: go
        output-file-path: bench.txt
        alert-threshold: "150%"  # Alert on 50% regression
```

## Acceptance Criteria

- [ ] All 8 adapters have `adapter_bench_test.go` with standardized benchmarks
- [ ] `BenchmarkFixture` generator exists in `internal/adapter/testutil/`
- [ ] Fixtures cover small (10KB), medium (500KB), large (10MB), huge (100MB) sizes
- [ ] Benchmarks include `-benchmem` allocation tracking
- [ ] `go test -bench=. ./internal/adapter/...` runs all benchmarks
- [ ] CI runs benchmarks on PRs with regression detection (>50% alert)
- [ ] Benchmark results are comparable across adapters (same metric names)
- [ ] SQLite-based adapters use in-memory databases for benchmarks (no disk I/O variance)

## Dependencies

- `td-2002fa` (shared adapter utils) — for consistent test fixture helpers

## Risks

- **Low**: SQLite benchmarks may be noisy due to I/O
- **Mitigation**: Use `:memory:` SQLite databases for benchmarks; use b.ResetTimer() after setup

## Scenarios

### Scenario: Detect performance regression
```
Given current BenchmarkMessages_Large/claudecode = 2.3ms
When a PR changes the Claude Code parser
And benchmark shows 5.0ms (>150% regression)
Then CI flags the regression in the PR check
```

### Scenario: Compare adapters
```
Given all 8 adapters have BenchmarkMessages_Large
When benchmarks are run
Then a developer can see which adapter is slowest
And prioritize optimization work (e.g., td-0da8fe)
```

### Scenario: Validate optimization
```
Given BenchmarkMessages_Large/geminicli = 8.0ms before optimization
When incremental parsing is implemented (td-0da8fe)
And benchmark shows 1.5ms
Then the 80% improvement target is confirmed
```
