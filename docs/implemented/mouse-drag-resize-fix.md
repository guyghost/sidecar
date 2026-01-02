# Mouse Drag-to-Resize Fix

## Problem
Pane drag-to-resize was implemented but never worked. Users could see the visible divider but clicking/dragging on it had no effect.

## Root Cause
Two coordinate calculation bugs in the hit region registration:

### 1. Hit Region X Position Off-by-One
The divider hit region was placed at `treeWidth - 1`, but:
- Left pane with `Width(treeWidth)` occupies columns `0` to `treeWidth-1`
- The visual divider is rendered at column `treeWidth`
- So the hit region should start at `treeWidth`, not `treeWidth - 1`

**Wrong:**
```go
dividerX := p.treeWidth
p.mouseHandler.HitMap.AddRect(regionPaneDivider, dividerX-1, paneY, 3, paneHeight, nil)
```

**Correct:**
```go
dividerX := p.treeWidth
p.mouseHandler.HitMap.AddRect(regionPaneDivider, dividerX, paneY, 3, paneHeight, nil)
```

### 2. Hit Region Priority (Registration Order)
The HitMap tests regions in reverse order of registration (last added = highest priority). The divider was registered BEFORE the preview pane, so clicks on the overlapping area went to the preview pane instead of the divider.

**Wrong order:**
```go
p.mouseHandler.HitMap.AddRect(regionTreePane, ...)
p.mouseHandler.HitMap.AddRect(regionPaneDivider, ...)  // Lower priority
p.mouseHandler.HitMap.AddRect(regionPreviewPane, ...)  // Higher priority - wins!
```

**Correct order:**
```go
p.mouseHandler.HitMap.AddRect(regionTreePane, ...)      // Lowest priority
p.mouseHandler.HitMap.AddRect(regionPreviewPane, ...)   // Medium priority
p.mouseHandler.HitMap.AddRect(regionPaneDivider, ...)   // Highest priority - wins!
```

### 3. Width Reset on Every Render
`calculatePaneWidths()` was called every render and always reset `treeWidth` to the default 30%. Any drag changes were immediately overwritten.

**Fix:** Only set default width if `treeWidth == 0` (not yet initialized).

## Debugging Approach
Added logging at multiple levels to trace mouse events:
1. App level: `APP MOUSE: x=... y=... action=... button=...`
2. Plugin handleMouse: `MOUSE EVENT` and `MOUSE ACTION`
3. Click handler: `CLICK x=... y=... region=...`
4. Hit region registration: `DIVIDER HIT REGION x=... w=...`

This revealed that clicks were being received (`action=0` Press events) but the region was being identified as `preview-pane` instead of `pane-divider` due to the coordinate/priority issues.

## Files Changed
- `internal/plugins/filebrowser/view.go` - Fixed hit region coordinates and registration order
- `internal/plugins/filebrowser/mouse.go` - Fixed width calculation in drag handler
- `internal/plugins/gitstatus/sidebar_view.go` - Same fixes
- `internal/plugins/gitstatus/mouse.go` - Same fixes

## Persistence
Pane widths are now persisted to `~/.config/sidecar/state.json`:
- Load saved width in plugin `Init()` using `state.GetFileBrowserTreeWidth()` / `state.GetGitStatusSidebarWidth()`
- Save width on drag end using `state.SetFileBrowserTreeWidth()` / `state.SetGitStatusSidebarWidth()`

## Key Lesson
When debugging mouse hit regions:
1. Log the exact coordinates of clicks and hit regions
2. Verify hit region registration order (reverse order = priority)
3. Account for lipgloss `Width()` setting total width including borders
