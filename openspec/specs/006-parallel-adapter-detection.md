# OpenSpec: Parallelize Adapter Detection at Startup

**ID**: `td-1e5298`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Performance
**Effort**: 3 story points

## Problem Statement

`DetectAdapters()` in `internal/adapter/detect.go` calls each adapter's `Detect(projectRoot)` sequentially. With 8 adapters, each potentially scanning directories and reading files, startup detection can take 500ms-2s depending on disk I/O and number of AI tools installed. Some adapters (Claude Code, Codex, OpenCode) scan large directory trees in their `Detect()` method.

## Objective

Run adapter detection concurrently and add a persisted adapter index cache to skip full detection on subsequent launches for the same project.

## Constraints

- Detection results must be identical to sequential detection
- Adapter factory `init()` registration order must not affect results
- Cache must be invalidated when adapter data directories change
- Total detection time should be < 200ms for cached projects

## Technical Design

### Phase 1: Concurrent Detection

```go
func DetectAdapters(projectRoot string, logger *slog.Logger) []Adapter {
    factories := registeredFactories() // thread-safe copy
    
    type result struct {
        adapter Adapter
        err     error
    }
    
    results := make([]result, len(factories))
    var g errgroup.Group
    
    for i, factory := range factories {
        i, factory := i, factory
        g.Go(func() error {
            a := factory.Create(logger)
            detected, err := a.Detect(projectRoot)
            if detected && err == nil {
                results[i] = result{adapter: a}
            }
            return nil // never fail the group
        })
    }
    g.Wait()
    
    // Collect non-nil adapters, preserving registration order
    var adapters []Adapter
    for _, r := range results {
        if r.adapter != nil {
            adapters = append(adapters, r.adapter)
        }
    }
    return adapters
}
```

### Phase 2: Detection Index Cache

```go
// internal/adapter/cache/detection_cache.go

type DetectionIndex struct {
    ProjectRoot string            `json:"project_root"`
    Adapters    []DetectedAdapter `json:"adapters"`
    Timestamp   time.Time         `json:"timestamp"`
}

type DetectedAdapter struct {
    ID       string    `json:"id"`
    DataPath string    `json:"data_path"` // watched for invalidation
    ModTime  time.Time `json:"mod_time"`
}
```

Cache stored at `~/.config/sidecar/adapter-cache.json`.

### Invalidation Strategy

1. On startup, load cache for current `projectRoot`
2. For each cached adapter, stat the `DataPath` — if modtime changed, mark stale
3. If all adapters are fresh, use cached list (skip detection)
4. If any are stale, run full concurrent detection and update cache
5. Cache TTL: 24 hours max (force re-detect daily)

## Acceptance Criteria

- [ ] `DetectAdapters` runs all 8 adapters concurrently
- [ ] Adapter order in results matches registration order (deterministic)
- [ ] Startup time with cold cache: < 500ms (was 500ms-2s)
- [ ] Startup time with warm cache: < 200ms
- [ ] Cache is invalidated when adapter data directories change
- [ ] `go test -race ./internal/adapter/...` passes
- [ ] Manual cache clear: `sidecar --clear-adapter-cache`

## Dependencies

- None

## Risks

- **Low**: Some adapters may not be safe for concurrent `Detect()` (shared state in `init()`)
- **Mitigation**: Audit each adapter's `Detect()` for shared mutable state

## Scenarios

### Scenario: First launch on a project
```
Given no adapter cache exists for "/Users/dev/myproject"
When sidecar starts
Then all 8 adapters are detected concurrently
And found adapters are cached to disk
And total detection time is < 500ms
```

### Scenario: Subsequent launch with unchanged data
```
Given cache exists with Claude Code and Codex detected
When sidecar starts and both data directories have same modtime
Then cached adapter list is used (no Detect() calls)
And startup detection takes < 50ms
```

### Scenario: New AI tool installed
```
Given cache exists with only Claude Code detected
When the user installs Cursor and runs sidecar
Then cache detects Cursor data directory is new (not in cache)
And runs full concurrent detection
And cache is updated with Claude Code + Cursor
```
