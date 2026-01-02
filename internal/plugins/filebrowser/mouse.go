package filebrowser

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/mouse"
	"github.com/marcus/sidecar/internal/state"
)

// Mouse region identifiers
const (
	regionTreePane    = "tree-pane"    // Overall tree pane for scroll targeting
	regionPreviewPane = "preview-pane" // Overall preview pane for scroll targeting
	regionPaneDivider = "pane-divider" // Border between tree and preview
	regionTreeItem    = "tree-item"    // Individual file/folder (Data: visible index)
	regionQuickOpen   = "quick-open"   // Quick open modal item (Data: match index)
)

// handleMouse processes mouse events and dispatches to appropriate handlers.
func (p *Plugin) handleMouse(msg tea.MouseMsg) (*Plugin, tea.Cmd) {
	// Handle quick open modal first if active
	if p.quickOpenMode {
		return p.handleQuickOpenMouse(msg)
	}

	action := p.mouseHandler.HandleMouse(msg)

	switch action.Type {
	case mouse.ActionClick:
		return p.handleMouseClick(action)
	case mouse.ActionDoubleClick:
		return p.handleMouseDoubleClick(action)
	case mouse.ActionScrollUp, mouse.ActionScrollDown:
		return p.handleMouseScroll(action)
	case mouse.ActionDrag:
		return p.handleMouseDrag(action)
	case mouse.ActionDragEnd:
		return p.handleMouseDragEnd()
	}
	return p, nil
}

// handleMouseClick handles single click actions.
func (p *Plugin) handleMouseClick(action mouse.MouseAction) (*Plugin, tea.Cmd) {
	if action.Region == nil {
		return p, nil
	}

	switch action.Region.ID {
	case regionTreeItem:
		idx, ok := action.Region.Data.(int)
		if !ok {
			return p, nil
		}
		p.treeCursor = idx
		p.activePane = PaneTree
		p.ensureTreeCursorVisible()
		return p, p.loadPreviewForCursor()

	case regionTreePane:
		p.activePane = PaneTree
		return p, nil

	case regionPreviewPane:
		p.activePane = PanePreview
		return p, nil

	case regionPaneDivider:
		// Start drag with current tree width
		p.mouseHandler.StartDrag(action.X, action.Y, regionPaneDivider, p.treeWidth)
		return p, nil
	}

	return p, nil
}

// handleMouseDoubleClick handles double click actions.
func (p *Plugin) handleMouseDoubleClick(action mouse.MouseAction) (*Plugin, tea.Cmd) {
	if action.Region == nil || action.Region.ID != regionTreeItem {
		return p, nil
	}

	idx, ok := action.Region.Data.(int)
	if !ok {
		return p, nil
	}

	node := p.tree.GetNode(idx)
	if node == nil {
		return p, nil
	}

	if node.IsDir {
		// Toggle folder expand/collapse
		_ = p.tree.Toggle(node)
		p.treeCursor = idx
		p.ensureTreeCursorVisible()
		return p, nil
	}

	// Open file in editor (same as 'e' key)
	return p, p.openFile(node.Path)
}

// handleMouseScroll handles scroll wheel actions.
func (p *Plugin) handleMouseScroll(action mouse.MouseAction) (*Plugin, tea.Cmd) {
	// Determine which pane to scroll based on region or X position
	inTreePane := false
	if action.Region != nil {
		inTreePane = action.Region.ID == regionTreePane || action.Region.ID == regionTreeItem
	} else {
		inTreePane = action.X < p.treeWidth
	}

	delta := 3
	if action.Type == mouse.ActionScrollUp {
		delta = -3
	}

	if inTreePane {
		// Scroll tree by moving cursor
		p.treeCursor += delta
		if p.treeCursor < 0 {
			p.treeCursor = 0
		} else if p.treeCursor >= p.tree.Len() {
			p.treeCursor = p.tree.Len() - 1
		}
		p.ensureTreeCursorVisible()
		return p, p.loadPreviewForCursor()
	}

	// Scroll preview pane
	lines := p.previewHighlighted
	if len(lines) == 0 {
		lines = p.previewLines
	}
	visibleHeight := p.visibleContentHeight()
	maxScroll := len(lines) - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	p.previewScroll += delta
	if p.previewScroll < 0 {
		p.previewScroll = 0
	} else if p.previewScroll > maxScroll {
		p.previewScroll = maxScroll
	}

	return p, nil
}

// handleMouseDrag handles drag actions (pane resizing).
func (p *Plugin) handleMouseDrag(action mouse.MouseAction) (*Plugin, tea.Cmd) {
	if p.mouseHandler.DragRegion() != regionPaneDivider {
		return p, nil
	}

	startValue := p.mouseHandler.DragStartValue()
	newWidth := startValue + action.DragDX

	// Clamp to reasonable bounds (match calculatePaneWidths logic)
	available := p.width - 6 - dividerWidth
	minWidth := 20
	maxWidth := available - 40 // Leave at least 40 for preview
	if maxWidth < minWidth {
		maxWidth = minWidth
	}
	if newWidth < minWidth {
		newWidth = minWidth
	} else if newWidth > maxWidth {
		newWidth = maxWidth
	}

	p.treeWidth = newWidth
	p.previewWidth = available - p.treeWidth

	return p, nil
}

// handleMouseDragEnd handles the end of a drag operation (saves pane width).
func (p *Plugin) handleMouseDragEnd() (*Plugin, tea.Cmd) {
	// Save the current tree width to state
	_ = state.SetFileBrowserTreeWidth(p.treeWidth)
	return p, nil
}

// handleQuickOpenMouse handles mouse events in quick open modal.
func (p *Plugin) handleQuickOpenMouse(msg tea.MouseMsg) (*Plugin, tea.Cmd) {
	action := p.mouseHandler.HandleMouse(msg)

	switch action.Type {
	case mouse.ActionClick:
		if action.Region != nil && action.Region.ID == regionQuickOpen {
			if idx, ok := action.Region.Data.(int); ok {
				p.quickOpenCursor = idx
			}
		}
		return p, nil

	case mouse.ActionDoubleClick:
		if action.Region != nil && action.Region.ID == regionQuickOpen {
			if idx, ok := action.Region.Data.(int); ok {
				p.quickOpenCursor = idx
				plug, cmd := p.selectQuickOpenMatch()
				return plug.(*Plugin), cmd
			}
		}
		return p, nil

	case mouse.ActionScrollUp, mouse.ActionScrollDown:
		// Scroll quick open list
		delta := 3
		if action.Type == mouse.ActionScrollUp {
			delta = -3
		}
		p.quickOpenCursor += delta
		if p.quickOpenCursor < 0 {
			p.quickOpenCursor = 0
		} else if p.quickOpenCursor >= len(p.quickOpenMatches) {
			p.quickOpenCursor = len(p.quickOpenMatches) - 1
		}
		return p, nil
	}

	return p, nil
}
