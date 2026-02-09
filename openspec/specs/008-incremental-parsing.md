# OpenSpec: Add Incremental Parsing to All Adapters

**ID**: `td-0da8fe`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Performance
**Effort**: 5 story points

## Problem Statement

Claude Code and Codex adapters use incremental parsing (byte offset tracking, only parsing new data on append-only file growth). However, Gemini CLI and OpenCode adapters do full JSON parse on every refresh. For large sessions (>10MB), this creates noticeable latency:

- **Gemini CLI**: Reads entire `session-*.json` file, parses full JSON
- **OpenCode**: Reads all message files in `storage/messages/{session}/`, parses each fully

Additionally, Amp reads entire files just to check `env.initial.trees` for project matching.

## Objective

Implement incremental or delta-based parsing for Gemini CLI, OpenCode, and Amp adapters, reducing refresh latency by 80%+ for large sessions.

## Constraints

- Must produce identical results to full parse
- Must handle file truncation/rewrite gracefully (reset offset)
- Must work with the existing `cache.Cache[T]` infrastructure
- No changes to adapter interface

## Technical Design

### Strategy per Adapter

| Adapter | Current | Proposed Strategy |
|---------|---------|-------------------|
| Gemini CLI | Full JSON parse | File size + mtime cache key; skip parse if unchanged |
| OpenCode | Read all message files | Track known message file list; only parse new files |
| Amp | Read full thread file | File size + mtime for project match; offset for messages |

### Gemini CLI: Metadata-Based Caching

Gemini CLI sessions are single JSON files that grow. Strategy: cache parsed result keyed by (path, size, mtime).

```go
type cachedSession struct {
    Size     int64
    ModTime  time.Time
    Messages []Message
}

func (a *Adapter) Messages(sessionID string) ([]Message, error) {
    path := a.sessionPath(sessionID)
    stat, _ := os.Stat(path)
    
    if cached, ok := a.cache.Get(sessionID); ok {
        if cached.Size == stat.Size() && cached.ModTime == stat.ModTime() {
            return adapterutil.CopyMessages(cached.Messages), nil
        }
    }
    
    // Full parse only when file changed
    messages, err := a.parseSessionFile(path)
    if err != nil {
        return nil, err
    }
    
    a.cache.Set(sessionID, cachedSession{
        Size: stat.Size(), ModTime: stat.ModTime(), Messages: messages,
    })
    return messages, nil
}
```

### OpenCode: Incremental File Discovery

OpenCode stores each message as a separate file. Strategy: track known files, only parse new ones.

```go
type sessionCache struct {
    KnownFiles map[string]struct{} // filename → parsed
    Messages   []Message
}

func (a *Adapter) Messages(sessionID string) ([]Message, error) {
    cached := a.cache.GetOrCreate(sessionID)
    
    files, _ := os.ReadDir(a.messagesDir(sessionID))
    var newFiles []os.DirEntry
    for _, f := range files {
        if _, known := cached.KnownFiles[f.Name()]; !known {
            newFiles = append(newFiles, f)
        }
    }
    
    if len(newFiles) == 0 {
        return adapterutil.CopyMessages(cached.Messages), nil
    }
    
    // Only parse new files
    for _, f := range newFiles {
        msg, err := a.parseMessageFile(filepath.Join(a.messagesDir(sessionID), f.Name()))
        if err == nil {
            cached.Messages = append(cached.Messages, msg)
            cached.KnownFiles[f.Name()] = struct{}{}
        }
    }
    
    sort.Slice(cached.Messages, func(i, j int) bool {
        return cached.Messages[i].Timestamp.Before(cached.Messages[j].Timestamp)
    })
    
    return adapterutil.CopyMessages(cached.Messages), nil
}
```

### Amp: Offset-Based Thread Reading

Amp threads are JSON files that grow. Use file size tracking to detect changes.

```go
func (a *Adapter) threadMatchesProject(path, projectRoot string) (bool, error) {
    stat, _ := os.Stat(path)
    if cached, ok := a.matchCache.Get(path); ok {
        if cached.Size == stat.Size() {
            return cached.Matches, nil
        }
    }
    // Only read env section (first 4KB typically sufficient)
    // instead of reading entire file
    matches := a.checkProjectMatchFromHeader(path, projectRoot)
    a.matchCache.Set(path, matchResult{Size: stat.Size(), Matches: matches})
    return matches, nil
}
```

## Acceptance Criteria

- [ ] Gemini CLI: Messages() skips parse when file unchanged (size+mtime match)
- [ ] OpenCode: Messages() only parses new message files (not re-parsing known ones)
- [ ] Amp: threadMatchesProject() caches match result, reads only header for new checks
- [ ] All adapter tests pass with identical results
- [ ] Benchmark: 80%+ reduction in refresh latency for sessions with >100 messages
- [ ] File truncation/rewrite correctly resets cache (detected via size decrease)

## Dependencies

- `td-2002fa` (shared adapter utils) — for `adapterutil.CopyMessages`

## Risks

- **Medium**: OpenCode may modify existing message files (not just add new ones)
- **Mitigation**: Include mtime check per file, not just filename tracking

## Scenarios

### Scenario: Gemini CLI session with 500 messages
```
Given a session file with 500 messages (5MB)
When Messages() is called a second time with no file changes
Then cached result is returned in < 1ms
And no JSON parsing occurs
```

### Scenario: OpenCode session receives new message
```
Given a cached session with 50 known message files
When a new message file appears in the directory
Then only the new file is parsed
And it is appended to the cached message list
And total refresh time is proportional to 1 file, not 51
```

### Scenario: File rewrite resets cache
```
Given a cached Gemini CLI session (size=5MB)
When the AI tool rewrites the file (size drops to 2MB)
Then the cache detects size decrease
And performs a full re-parse
```
