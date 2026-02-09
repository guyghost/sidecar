# OpenSpec: Immutable Styles System — Thread Safety

**ID**: `td-396c71`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P1
**Type**: Refactoring
**Effort**: 5 story points

## Problem Statement

`internal/styles/styles.go` defines ~50 package-level mutable `var` color values and ~30 mutable style variables. `ApplyThemeColors()` in `themes.go` explicitly warns it is "NOT thread-safe" and mutates all these globals. Additionally, `rebuildStyles()` in `themes.go` (~1188 lines) manually duplicates every style definition from `styles.go`, creating a maintenance hazard where adding a new style requires updating two files.

## Objective

Replace the mutable global state with an immutable `Palette` + `Styles` struct system. Eliminate the duplication between `styles.go` and `themes.go`. Make theme switching thread-safe.

## Constraints

- All existing color references (`styles.Primary`, `styles.TextPrimary`, etc.) must continue compiling
- Theme switching must remain < 16ms (single frame)
- All 6 built-in themes + 453 community themes must work
- No visible rendering changes

## Technical Design

### Phase 1: Define Palette Struct

```go
// Palette holds all color values for a theme.
type Palette struct {
    Primary, Secondary, Accent         lipgloss.Color
    Success, Warning, Error, Info      lipgloss.Color
    TextPrimary, TextSecondary         lipgloss.Color
    TextMuted, TextSubtle              lipgloss.Color
    BgPrimary, BgSecondary, BgTertiary lipgloss.Color
    BorderNormal, BorderActive         lipgloss.Color
    DiffAddFg, DiffRemoveFg            lipgloss.Color
    // ... all current color vars
}

func DefaultPalette() Palette { ... }
```

### Phase 2: Define Styles Struct

```go
// Styles holds all pre-computed lipgloss styles derived from a Palette.
type Styles struct {
    palette Palette
    // All current style vars become fields
    Header, Footer         lipgloss.Style
    TabActive, TabInactive lipgloss.Style
    DiffAdd, DiffRemove    lipgloss.Style
    // ... etc
}

// NewStyles creates an immutable Styles from a Palette. 
// This is the SINGLE place where styles are computed from colors.
func NewStyles(p Palette) *Styles { ... }
```

### Phase 3: Thread-safe Theme Manager

```go
// ThemeManager provides atomic access to the current styles.
type ThemeManager struct {
    current atomic.Pointer[Styles]
}

func (tm *ThemeManager) Current() *Styles
func (tm *ThemeManager) Apply(p Palette)
```

### Phase 4: Migration Path

Keep package-level accessor functions for backward compatibility:

```go
// Deprecated: Use ThemeManager.Current().Header instead.
// These exist only for migration and will be removed.
func CurrentStyles() *Styles { return manager.Current() }
```

Or use a global `var Current *Styles` that is atomically swapped on theme change (safe for concurrent reads after swap).

### Phase 5: Eliminate Duplication

Delete `rebuildStyles()` entirely. `NewStyles()` is now the single source of truth. Theme overrides modify the `Palette` before passing to `NewStyles()`.

## Acceptance Criteria

- [ ] Zero mutable package-level `var` for colors or styles
- [ ] Single `NewStyles(Palette)` constructor (no `rebuildStyles` duplication)
- [ ] `ApplyThemeColors` replaced with thread-safe `ThemeManager.Apply()`
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] All 6 built-in themes render identically (visual comparison)
- [ ] Community theme application works
- [ ] Theme switching performance < 16ms

## Dependencies

- Should be done BEFORE `001-decompose-app-model` (Model references styles extensively)

## Risks

- **Medium**: Every file in the project references `styles.XYZ` — massive find-replace
- **Mitigation**: Phase 4 provides backward-compatible accessors during migration

## Scenarios

### Scenario: Theme switch is thread-safe
```
Given theme "Dracula" is active
When the user switches to "Nord" theme
Then all concurrent View() calls see either full Dracula or full Nord
And no partial theme state is visible
```

### Scenario: New style only defined once
```
Given a developer adds a new style "CodeBlockBg"
When they define it in NewStyles()
Then it is available for all themes automatically
And no second definition is needed
```

### Scenario: Community theme with missing overrides
```
Given a community theme only overrides 5 of 50 colors
When the theme is applied
Then the missing colors use DefaultPalette() values
And all styles are correctly derived
```
