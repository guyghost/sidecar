# OpenSpec: Extract Shared Adapter Utilities

**ID**: `td-2002fa`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P1
**Type**: Refactoring
**Effort**: 3 story points

## Problem Statement

Multiple utility functions are duplicated across 4+ adapter implementations:

| Function | Duplicated In |
|----------|--------------|
| `copyMessages([]Message)` | claudecode, codex, amp, opencode |
| `shortID(string)` | claudecode, codex, geminicli, amp, opencode |
| `truncateTitle(string, int)` | claudecode, codex, cursor, warp |
| `cwdMatchesProject(string, string)` | claudecode, codex, amp, opencode |
| `resolveProjectPath(string)` | claudecode, codex, geminicli |
| Cost estimation with hardcoded model prices | claudecode, geminicli, opencode |

This creates maintenance burden and inconsistency (e.g., different `shortID` prefix lengths across adapters).

## Objective

Extract all duplicated adapter utilities into `internal/adapter/adapterutil/` with consistent implementations and a configurable pricing table.

## Constraints

- Existing adapter behavior must not change
- All adapter tests must continue passing
- No new external dependencies

## Technical Design

### Package Structure

```
internal/adapter/adapterutil/
├── messages.go      // CopyMessages, SortMessagesByTime
├── strings.go       // ShortID, TruncateTitle, SanitizeTitle
├── paths.go         // CWDMatchesProject, ResolveProjectPath, ExpandHome
├── pricing.go       // PricingTable, EstimateCost
└── pricing_test.go
```

### Key APIs

```go
package adapterutil

// CopyMessages returns a deep copy of a message slice.
func CopyMessages(msgs []Message) []Message

// ShortID returns the first 8 characters of an ID string.
func ShortID(id string, length ...int) string

// TruncateTitle truncates a title to maxLen, adding "..." if truncated.
func TruncateTitle(title string, maxLen int) string

// CWDMatchesProject checks if a working directory matches or is under a project root.
func CWDMatchesProject(cwd, projectRoot string) bool

// ResolveProjectPath resolves ~, symlinks, and normalizes a project path.
func ResolveProjectPath(path string) string

// PricingTable holds per-model token pricing.
type PricingTable map[string]ModelPricing

type ModelPricing struct {
    InputPerMillion  float64
    OutputPerMillion float64
    CacheReadPerMillion  float64
    CacheWritePerMillion float64
}

// DefaultPricing returns the built-in pricing table.
func DefaultPricing() PricingTable

// EstimateCost calculates cost from token usage.
func (pt PricingTable) EstimateCost(model string, usage TokenUsage) float64
```

### Migration

For each adapter:
1. Replace internal function with `adapterutil.XYZ` call
2. Remove the local copy
3. Run adapter-specific tests

## Acceptance Criteria

- [ ] `internal/adapter/adapterutil/` package exists with all listed functions
- [ ] Zero duplicated utility functions remain in individual adapters
- [ ] `go test ./internal/adapter/...` passes for all 8 adapters
- [ ] `ShortID` behavior is consistent across all adapters (standardized to 8 chars)
- [ ] `PricingTable` is configurable (not hardcoded per adapter)
- [ ] 100% test coverage on `adapterutil` package

## Dependencies

- None

## Risks

- **Low**: Subtle behavioral differences in `cwdMatchesProject` between adapters
- **Mitigation**: Catalog each adapter's implementation, test with all existing test fixtures

## Scenarios

### Scenario: Consistent short IDs across adapters
```
Given a session with ID "abc123def456"
When Claude Code adapter calls ShortID
And Codex adapter calls ShortID
Then both return "abc123de" (8 chars)
```

### Scenario: Cost estimation with custom pricing
```
Given Claude 3.5 Sonnet usage: 1000 input tokens, 500 output tokens
When EstimateCost is called with DefaultPricing
Then the cost is calculated from the centralized pricing table
And updating the price only requires changing DefaultPricing()
```

### Scenario: Path matching with symlinks
```
Given project root is "/Users/dev/project" which is a symlink to "/opt/projects/myapp"
When CWDMatchesProject("/opt/projects/myapp/src", "/Users/dev/project") is called
Then it returns true (resolves symlinks before comparison)
```
