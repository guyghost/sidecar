# OpenSpec: Performance Metrics and Diagnostics Panel

**ID**: `td-9d1446`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P3
**Type**: Feature
**Effort**: 5 story points

## Problem Statement

Sidecar has pprof support via `SIDECAR_PPROF` env var, but no lightweight runtime metrics for production use. When users report sluggishness, there's no way to see render times, event drops, cache hit rates, or adapter latencies without attaching a profiler. The existing diagnostics modal (`ModalDiagnostics`) shows version info but no performance data.

## Objective

Add lightweight performance counters collected at runtime, accessible via the existing diagnostics panel. Counters should have near-zero overhead when not being read.

## Constraints

- Metric collection overhead < 1μs per data point
- No external dependencies (no Prometheus, no OpenTelemetry)
- Must work without pprof enabled
- Must not allocate on the hot path
- Metrics reset on project switch

## Technical Design

### Metrics Registry

```go
// internal/metrics/metrics.go

type Counter struct {
    value atomic.Int64
}

func (c *Counter) Inc()
func (c *Counter) Add(n int64)
func (c *Counter) Value() int64

type Gauge struct {
    value atomic.Int64
}

func (g *Gauge) Set(n int64)
func (g *Gauge) Value() int64

type Histogram struct {
    // Lock-free approximate histogram using atomic buckets
    buckets [16]atomic.Int64 // log2 buckets: <1ms, <2ms, <4ms, ..., <32s
    count   atomic.Int64
    sum     atomic.Int64
}

func (h *Histogram) Observe(d time.Duration)
func (h *Histogram) P50() time.Duration
func (h *Histogram) P99() time.Duration
func (h *Histogram) Count() int64
```

### Pre-defined Metrics

```go
var (
    // Rendering
    RenderDuration  = NewHistogram("render_duration")
    RenderCount     = NewCounter("render_count")
    
    // Event bus
    EventsPublished = NewCounter("events_published")
    EventsDropped   = NewCounter("events_dropped")
    
    // Adapter
    AdapterRefreshDuration = NewHistogramVec("adapter_refresh_duration", "adapter_id")
    AdapterCacheHits       = NewCounterVec("adapter_cache_hits", "adapter_id")
    AdapterCacheMisses     = NewCounterVec("adapter_cache_misses", "adapter_id")
    
    // Git
    GitOperationDuration = NewHistogramVec("git_op_duration", "operation")
    
    // Memory
    GoroutineCount = NewGauge("goroutine_count")
    HeapAlloc      = NewGauge("heap_alloc_bytes")
)
```

### Diagnostics Panel Integration

Extend the existing `ModalDiagnostics` to include a "Performance" tab:

```
┌─────────────────── Diagnostics ───────────────────┐
│ [Version] [Performance] [System]                   │
│                                                    │
│ Render:  avg=2.1ms  p99=8.3ms  count=1,204       │
│ Events:  published=3,421  dropped=0               │
│                                                    │
│ Adapters:                                          │
│   claude-code: refresh=12ms  cache=94% hit        │
│   codex:       refresh=8ms   cache=87% hit        │
│                                                    │
│ Git:                                               │
│   status=15ms  diff=22ms  log=45ms               │
│                                                    │
│ System:                                            │
│   goroutines=34  heap=12.3MB  FDs=28             │
└────────────────────────────────────────────────────┘
```

### Collection Points

| Where | What | How |
|-------|------|-----|
| `app/view.go` `View()` | Render duration | `defer RenderDuration.Observe(time.Since(start))` |
| `event/dispatcher.go` `Publish()` | Event drops | `EventsDropped.Inc()` on channel full |
| Each adapter `Messages()` | Refresh latency | Wrap in histogram |
| Each adapter cache | Hit/miss | Counter in `cache.Get()` |
| `internal/git/` operations | Git latency | Wrap in histogram |
| Periodic tick (5s) | Goroutines, heap | `runtime.NumGoroutine()`, `runtime.ReadMemStats()` |

## Acceptance Criteria

- [ ] `internal/metrics/` package exists with Counter, Gauge, Histogram types
- [ ] All types use atomics only (no mutexes on hot path)
- [ ] Render duration is tracked in `app/view.go`
- [ ] Event drops are tracked in `event/dispatcher.go`
- [ ] Adapter refresh latency and cache rates are tracked
- [ ] Diagnostics panel shows performance tab
- [ ] `go test -bench ./internal/metrics/` shows < 100ns per metric operation
- [ ] Metrics reset on project switch
- [ ] Feature-flagged: `metrics_panel` (default enabled)

## Dependencies

- `td-3bf96f` (event bus improvements) — for adding drop counter
- Beneficial after `td-7e3911` (extract git package) — for git operation metrics

## Risks

- **Low**: Atomic operations on ARM64 may have different overhead
- **Mitigation**: Benchmark on both darwin/arm64 and linux/amd64

## Scenarios

### Scenario: User reports sluggish UI
```
Given sidecar feels slow
When the user presses "!" to open diagnostics
And selects the "Performance" tab
Then they see render p99=45ms (indicating slow renders)
And adapter claude-code refresh=200ms (identifying the bottleneck)
```

### Scenario: Zero overhead when not reading
```
Given metrics are being collected
When no diagnostics panel is open
Then the metric collection overhead is < 1μs per View() call
And no allocations occur from metrics
```
