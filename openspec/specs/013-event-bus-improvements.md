# OpenSpec: Event Bus Improvements — Buffer Sizing and Backpressure

**ID**: `td-3bf96f`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Enhancement
**Effort**: 3 story points

## Problem Statement

The event dispatcher in `internal/event/dispatcher.go` uses a fixed buffer size of 16 per subscriber channel. When a subscriber is slow (e.g., during a heavy render cycle), events are silently dropped with only a warning log. There are no metrics on drops, no backpressure mechanism, and no way to configure buffer sizes per topic.

In high-activity scenarios (rapid file changes during a git rebase, multiple adapter watch events), drops can cause missed updates visible as stale UI.

## Objective

Increase default buffer, add per-topic configuration, add drop metrics, and implement optional backpressure for critical topics.

## Constraints

- Must remain lock-free on the publish path (non-blocking)
- Must not break existing subscriber patterns
- Bubble Tea's single-threaded update loop must not be blocked

## Technical Design

### Configurable Buffer Sizes

```go
const (
    DefaultBufferSize  = 64   // was 16
    CriticalBufferSize = 128  // for topics that must not drop
)

type SubscribeOption func(*subscribeConfig)

func WithBufferSize(size int) SubscribeOption {
    return func(c *subscribeConfig) { c.bufferSize = size }
}

func (d *Dispatcher) Subscribe(topic string, opts ...SubscribeOption) <-chan Event {
    cfg := &subscribeConfig{bufferSize: DefaultBufferSize}
    for _, opt := range opts {
        opt(cfg)
    }
    ch := make(chan Event, cfg.bufferSize)
    // ...
}
```

### Drop Metrics

```go
type Dispatcher struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
    closed      bool
    logger      *slog.Logger
    
    // Metrics
    published atomic.Int64
    dropped   atomic.Int64
}

func (d *Dispatcher) Publish(topic string, event Event) {
    d.published.Add(1)
    // ...
    select {
    case ch <- event:
    default:
        d.dropped.Add(1)
        d.logger.Warn("event dropped", "topic", topic, ...)
    }
}

func (d *Dispatcher) Stats() DispatcherStats {
    return DispatcherStats{
        Published: d.published.Load(),
        Dropped:   d.dropped.Load(),
    }
}
```

### Overflow Strategy per Topic

```go
type OverflowStrategy int

const (
    OverflowDrop    OverflowStrategy = iota // Current behavior (non-blocking)
    OverflowDropOldest                       // Drop oldest event in buffer, add new
)

func WithOverflow(strategy OverflowStrategy) SubscribeOption
```

`OverflowDropOldest` implementation:
```go
case OverflowDropOldest:
    select {
    case ch <- event:
    default:
        // Drain oldest, push new
        select {
        case <-ch:
            d.dropped.Add(1)
        default:
        }
        ch <- event
    }
```

### Topic Naming Convention

Define standard topic constants:

```go
const (
    TopicGitChange     = "git.change"
    TopicAdapterWatch  = "adapter.watch"
    TopicConfigChange  = "config.change"
    TopicProjectSwitch = "project.switch"
)
```

## Acceptance Criteria

- [ ] Default buffer size increased from 16 to 64
- [ ] `Subscribe()` accepts `WithBufferSize()` option
- [ ] `Subscribe()` accepts `WithOverflow()` option
- [ ] `Stats()` returns published and dropped event counts
- [ ] `OverflowDropOldest` strategy works correctly
- [ ] Existing callers of `Subscribe(topic)` work without changes (backward compatible)
- [ ] `go test -race ./internal/event/...` passes
- [ ] Standard topic constants defined
- [ ] Drop metrics integrated with diagnostics panel (if `td-9d1446` is done)

## Dependencies

- None (but feeds into `td-9d1446` performance metrics)

## Risks

- **Low**: `OverflowDropOldest` adds a brief lock contention window
- **Mitigation**: Only use for topics where ordering matters; default remains `OverflowDrop`

## Scenarios

### Scenario: High-frequency git changes
```
Given a subscriber on "git.change" with buffer size 64
When 100 rapid file change events fire (during git rebase)
Then at most 36 are dropped (64 buffered)
And Stats() reports drops accurately
And the UI shows the final state correctly
```

### Scenario: DropOldest for adapter watch
```
Given a subscriber on "adapter.watch" with OverflowDropOldest
When the buffer is full (64 events)
And a new event arrives
Then the oldest event is removed
And the new event is added to the buffer
And the subscriber always sees the most recent events
```

### Scenario: Backward compatibility
```
Given existing code calling d.Subscribe("git.change")
When the new version is deployed
Then the subscriber gets a 64-element buffer (was 16)
And no code changes are required
```
