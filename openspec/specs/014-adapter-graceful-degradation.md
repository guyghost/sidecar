# OpenSpec: Adapter Graceful Degradation for Corrupt Files

**ID**: `td-73f37d`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Enhancement
**Effort**: 5 story points

## Problem Statement

No adapter handles corrupt or partially-written files gracefully. When an AI tool is mid-write (e.g., Claude Code appending to a JSONL file), the adapter may encounter:
- Truncated JSON at end of file
- Partial UTF-8 sequences
- Incomplete JSONL lines
- SQLite WAL checkpoint in progress

In all these cases, the adapter returns an error and the UI shows no data — rather than showing the successfully-parsed portion.

## Objective

Add recovery points in all adapter parsers so they return partial data with error metadata instead of failing entirely. The UI should show available data with a "partial data" indicator.

## Constraints

- Must not mask genuine errors (e.g., permission denied, file not found)
- Partial data must be clearly flagged (not silently incomplete)
- Recovery must not corrupt internal cache state
- Parse errors must be logged for debugging

## Technical Design

### Error Types

```go
// internal/adapter/errors.go

// PartialResult wraps adapter results that may be incomplete.
type PartialResult struct {
    Error      error   // The parse error encountered
    ParsedUpTo int64   // Byte offset of last successful parse
    Recovered  bool    // True if partial data was recovered
}

// IsPartial checks if an error contains partial result metadata.
func IsPartial(err error) (*PartialResult, bool) {
    var pr *PartialResult
    if errors.As(err, &pr) {
        return pr, true
    }
    return nil, false
}
```

### Adapter Behavior

```go
// Example: Claude Code JSONL parser
func (a *Adapter) Messages(sessionID string) ([]Message, error) {
    messages, bytesRead, err := a.parseMessagesIncremental(sessionID)
    
    if err != nil && len(messages) > 0 {
        // Partial success — return data + wrapped error
        return messages, &PartialResult{
            Error:      err,
            ParsedUpTo: bytesRead,
            Recovered:  true,
        }
    }
    
    if err != nil {
        // Total failure — no data recovered
        return nil, err
    }
    
    return messages, nil
}
```

### JSONL Recovery Strategy

For JSONL-based adapters (Claude Code, Codex):
1. Parse line by line
2. On malformed line: skip it, log warning, continue
3. On truncated last line: ignore it (likely mid-write)
4. Return all successfully parsed lines + PartialResult error

### JSON Recovery Strategy

For JSON-based adapters (Gemini CLI, OpenCode, Amp):
1. Attempt full parse
2. On error: try parsing up to the last complete JSON object/array element
3. For OpenCode (per-file messages): skip corrupt files, return rest

### SQLite Recovery Strategy

For SQLite-based adapters (Warp, Kiro, Cursor):
1. Set busy_timeout pragma (5 seconds)
2. On lock timeout: return cached data + PartialResult
3. On corrupt database: return empty + non-partial error (genuine failure)

### UI Integration

The conversations plugin shows a subtle indicator when displaying partial data:

```
┌─ Session: claude-code-abc123 ──────────────────────┐
│ ⚠ Partial data (conversation may be updating)       │
│                                                      │
│ [Message 1] ...                                      │
│ [Message 2] ...                                      │
│ [Message 3] ...                                      │
└──────────────────────────────────────────────────────┘
```

## Acceptance Criteria

- [ ] `PartialResult` error type defined in `internal/adapter/errors.go`
- [ ] Claude Code adapter recovers partial JSONL (truncated last line)
- [ ] Codex adapter recovers partial JSONL
- [ ] Gemini CLI adapter handles truncated JSON gracefully
- [ ] OpenCode adapter skips corrupt message files, returns rest
- [ ] Warp/Kiro/Cursor adapters handle SQLite busy with cached fallback
- [ ] Conversations plugin shows "partial data" indicator for `PartialResult`
- [ ] Genuine errors (permission denied, not found) are NOT wrapped as partial
- [ ] Parse errors are logged at WARN level with file path and offset
- [ ] `go test ./internal/adapter/...` includes corrupt file test cases

## Dependencies

- None

## Risks

- **Medium**: Defining "genuine error" vs "recoverable error" boundaries for each adapter
- **Mitigation**: Whitelist recoverable errors (EOF, truncated JSON) rather than blacklist

## Scenarios

### Scenario: Claude Code mid-write
```
Given Claude Code is actively writing a response (JSONL append in progress)
When sidecar reads the session file
Then 99 of 100 lines parse successfully
And the truncated 100th line is skipped
And Messages() returns 99 messages + PartialResult error
And the UI shows 99 messages with "⚠ updating" indicator
```

### Scenario: Permission denied is not partial
```
Given a session file with chmod 000
When the adapter reads it
Then it returns nil, error (permission denied)
And the error is NOT a PartialResult
And the UI shows an error state
```

### Scenario: SQLite busy during WAL checkpoint
```
Given Warp's SQLite is mid-checkpoint
When the adapter queries it
Then it retries for up to 5 seconds (busy_timeout)
If still locked, returns cached data + PartialResult
And the UI shows cached data with "⚠ database busy" indicator
```
