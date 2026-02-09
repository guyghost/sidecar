# OpenSpec: Unify SQLite Dependency to Pure Go

**ID**: `td-9a9f85`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Chore
**Effort**: 3 story points

## Problem Statement

The project uses two different SQLite libraries:
- `mattn/go-sqlite3` (CGo-based) — used by **Warp** and **Kiro** adapters
- `modernc.org/sqlite` (pure Go) — used by **Cursor** adapter and **td** integration

The release build uses `CGO_ENABLED=0`, making CGo SQLite incompatible with release binaries. Having both as dependencies increases binary size and creates confusion about which to use.

## Objective

Migrate all SQLite usage to `modernc.org/sqlite` (pure Go), remove `mattn/go-sqlite3` from `go.mod`, and ensure consistent database access patterns.

## Constraints

- All SQLite queries must produce identical results
- WAL mode compatibility must be maintained
- No performance regression > 10% for typical query sizes
- Cross-platform: darwin (arm64/amd64) + linux (arm64/amd64)

## Technical Design

### Phase 1: Audit Current Usage

| Adapter | Library | DB File | Query Patterns |
|---------|---------|---------|---------------|
| Warp | mattn/go-sqlite3 | `~/Library/.../warp.sqlite` | SELECT from ai_queries, agent_conversations, blocks |
| Kiro | mattn/go-sqlite3 | `~/Library/.../data.sqlite3` | SELECT from conversations_v2 |
| Cursor | modernc.org/sqlite | `~/.cursor/.../store.db` | SELECT from meta, ItemTable |

### Phase 2: Migrate Warp Adapter

1. Replace `_ "github.com/mattn/go-sqlite3"` import with `_ "modernc.org/sqlite"`
2. Change `sql.Open("sqlite3", ...)` to `sql.Open("sqlite", ...)`
3. Verify WAL mode query compatibility
4. Run all Warp adapter tests

### Phase 3: Migrate Kiro Adapter

Same steps as Warp. Kiro uses simpler queries (single table).

### Phase 4: Remove mattn/go-sqlite3

1. `go mod tidy` after removing all imports
2. Verify `mattn/go-sqlite3` is no longer in `go.sum`
3. Verify `CGO_ENABLED=0 go build ./...` succeeds

### Phase 5: Create Shared SQLite Helper

```go
package sqliteutil

import (
    "database/sql"
    _ "modernc.org/sqlite"
)

// OpenReadOnly opens a SQLite database in read-only mode with WAL.
func OpenReadOnly(path string) (*sql.DB, error) {
    return sql.Open("sqlite", path+"?mode=ro&_journal_mode=WAL")
}
```

## Acceptance Criteria

- [ ] `mattn/go-sqlite3` removed from `go.mod` and `go.sum`
- [ ] `CGO_ENABLED=0 go build ./...` succeeds
- [ ] `go test ./internal/adapter/warp/...` passes
- [ ] `go test ./internal/adapter/kiro/...` passes
- [ ] `go test ./internal/adapter/cursor/...` passes (no regression)
- [ ] Warp adapter reads identical data from warp.sqlite
- [ ] Kiro adapter reads identical data from data.sqlite3
- [ ] Binary size does not increase by more than 5%

## Dependencies

- None

## Risks

- **Low**: `modernc.org/sqlite` may have different default pragmas
- **Mitigation**: Explicitly set pragmas (journal_mode, busy_timeout) in `OpenReadOnly`
