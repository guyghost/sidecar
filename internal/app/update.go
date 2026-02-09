package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/keymap"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/palette"
	"github.com/guyghost/sidecar/internal/plugin"
	"github.com/guyghost/sidecar/internal/state"
	"github.com/guyghost/sidecar/internal/version"
)

// isMouseEscapeSequence returns true if the key message appears to be
// an unparsed mouse escape sequence (SGR format: [<...M or [<...m)
func isMouseEscapeSequence(msg tea.KeyMsg) bool {
	s := msg.String()
	// SGR mouse sequences contain [< and end with M or m
	if strings.Contains(s, "[<") && (strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m")) {
		return true
	}
	// Check for semicolon-separated coordinate patterns typical of mouse sequences
	if strings.Contains(s, ";") && strings.ContainsAny(s, "0123456789") {
		if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m") {
			return true
		}
	}
	return false
}

// Update handles all messages and returns the updated model and commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return (&m).handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			// Prime worktree cache before first render
			m.refreshWorktreeCache()
		}
		m.ready = true
		// Reset diagnostics modal on resize (will be rebuilt on next render)
		if m.showDiagnostics {
			m.diagnosticsModalWidth = 0
		}
		// Forward adjusted WindowSizeMsg to all plugins
		// Plugins receive the content area size (minus header and footer)
		// Must match the height passed to Plugin.View() in view.go
		adjustedHeight := msg.Height - headerHeight - footerHeight
		adjustedMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: adjustedHeight,
		}
		plugins := m.registry.Plugins()
		var cmds []tea.Cmd
		for i, p := range plugins {
			newPlugin, cmd := p.Update(adjustedMsg)
			plugins[i] = newPlugin
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case tea.MouseMsg:
		// Route mouse events to active modal (priority order)
		switch m.activeModal() {
		case ModalPalette:
			var cmd tea.Cmd
			m.palette, cmd = m.palette.Update(msg)
			return m, cmd
		case ModalHelp:
			return m.handleHelpModalMouse(msg)
		case ModalUpdate:
			return m.handleUpdateModalMouse(msg)
		case ModalDiagnostics:
			return m.handleDiagnosticsModalMouse(msg)
		case ModalQuitConfirm:
			return m.handleQuitConfirmMouse(msg)
		case ModalProjectSwitcher:
			if m.project.AddMode {
				return m.handleProjectAddModalMouse(msg)
			}
			return m.handleProjectSwitcherMouse(msg)
		case ModalWorktreeSwitcher:
			return m.handleWorktreeSwitcherMouse(msg)
		case ModalThemeSwitcher:
			return m.handleThemeSwitcherMouse(msg)
		case ModalIssueInput:
			return m.handleIssueInputMouse(msg)
		case ModalIssuePreview:
			return m.handleIssuePreviewMouse(msg)
		}

		// Handle header tab clicks (Y < 2 means header area)
		if msg.Y < headerHeight && msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if start, end, ok := m.getRepoNameBounds(); ok && !m.intro.Active && msg.X >= start && msg.X < end {
				m.project.Show = true
				m.activeContext = "project-switcher"
				m.initProjectSwitcher()
				return m, nil
			}

			// Check if click is on worktree indicator
			if start, end, ok := m.getWorktreeIndicatorBounds(); ok && !m.intro.Active && msg.X >= start && msg.X < end {
				worktrees := GetWorktrees(m.ui.WorkDir)
				if len(worktrees) > 1 {
					m.worktree.Show = true
					m.activeContext = "worktree-switcher"
					m.initWorktreeSwitcher()
					return m, nil
				}
			}

			// Check if click is on a tab
			tabBounds := m.getTabBounds()
			for i, bounds := range tabBounds {
				if msg.X >= bounds.Start && msg.X < bounds.End {
					return m, m.SetActivePlugin(i)
				}
			}
			return m, nil
		}

		// Forward mouse events to active plugin with Y offset for app header (2 lines)
		if p := m.ActivePlugin(); p != nil {
			adjusted := tea.MouseMsg{
				X:      msg.X,
				Y:      msg.Y - headerHeight, // Offset for app header
				Button: msg.Button,
				Action: msg.Action,
				Ctrl:   msg.Ctrl,
				Alt:    msg.Alt,
				Shift:  msg.Shift,
			}
			newPlugin, cmd := p.Update(adjusted)
			plugins := m.registry.Plugins()
			if m.activePlugin < len(plugins) {
				plugins[m.activePlugin] = newPlugin
			}
			m.updateContext()
			return m, cmd
		}
		return m, nil

	case IntroTickMsg:
		if m.intro.Active {
			m.intro.Update(16 * time.Millisecond)
			// Keep ticking until logo done AND repo name fully faded in
			if !m.intro.Done || m.intro.RepoOpacity < 1.0 {
				return m, IntroTick()
			}
			// All animations complete - mark intro as inactive so header clicks work
			m.intro.Active = false
			return m, Refresh()
		}
		return m, nil

	case TickMsg:
		m.ui.UpdateClock()
		m.ui.ClearExpiredToast()
		m.ClearToast()
		// Eagerly refresh worktree cache (must happen in Update, not View, due to value receiver)
		m.refreshWorktreeCache()
		// Periodically check if current worktree still exists (every 10 seconds)
		m.worktree.CheckCounter++
		if m.worktree.CheckCounter >= 10 {
			m.worktree.CheckCounter = 0
			return m, tea.Batch(tickCmd(), checkWorktreeExists(m.ui.WorkDir))
		}
		return m, tickCmd()

	case UpdateSpinnerTickMsg:
		if m.update.InProgress {
			m.update.SpinnerFrame = (m.update.SpinnerFrame + 1) % 10
			return m, updateSpinnerTick()
		}
		return m, nil

	case ToastMsg:
		m.ShowToast(msg.Message, msg.Duration)
		m.statusIsError = msg.IsError
		return m, nil

	case RefreshMsg:
		m.ui.MarkRefresh()
		// Refresh active plugin
		if p := m.ActivePlugin(); p != nil {
			_, cmd := p.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		m.lastError = msg.Err
		m.ShowToast("Error: "+msg.Err.Error(), 5*time.Second)
		return m, nil

	case UpdateSuccessMsg:
		m.update.InProgress = false
		m.update.NeedsRestart = true
		// Set all phases to done
		m.update.PhaseStatus[PhaseCheckPrereqs] = "done"
		m.update.PhaseStatus[PhaseInstalling] = "done"
		m.update.PhaseStatus[PhaseVerifying] = "done"
		// Update modal state if modal is open
		if m.update.ModalState == UpdateModalProgress {
			m.update.ModalState = UpdateModalComplete
		}
		if msg.SidecarUpdated {
			m.updateAvailable = nil
		}
		if msg.TdUpdated && m.tdVersionInfo != nil {
			m.tdVersionInfo.HasUpdate = false
		}
		// Only show toast if modal is not open
		if m.update.ModalState == UpdateModalClosed {
			m.ShowToast("Update complete! Restart sidecar to use new version", 10*time.Second)
		}
		return m, nil

	case UpdateErrorMsg:
		m.update.InProgress = false
		m.update.Error = fmt.Sprintf("Failed to update %s: %s", msg.Step, msg.Err)
		// Mark current phase as error
		m.update.PhaseStatus[m.update.Phase] = "error"
		// Update modal state if modal is open
		if m.update.ModalState == UpdateModalProgress {
			m.update.ModalState = UpdateModalError
		}
		// Only show toast if modal is not open
		if m.update.ModalState == UpdateModalClosed {
			m.ShowToast("Update failed: "+msg.Err.Error(), 5*time.Second)
		}
		m.statusIsError = true
		return m, nil

	case UpdatePhaseChangeMsg:
		m.update.PhaseStatus[msg.Phase] = msg.Status
		if msg.Status == "running" {
			m.update.Phase = msg.Phase
		}
		return m, nil

	case UpdateElapsedTickMsg:
		// Continue timer if update is in progress
		if m.update.InProgress && m.update.ModalState == UpdateModalProgress {
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return UpdateElapsedTickMsg{}
			})
		}
		return m, nil

	case UpdatePrereqsPassedMsg:
		// Prerequisites passed - transition to install phase
		m.update.PhaseStatus[PhaseCheckPrereqs] = "done"
		m.update.Phase = PhaseInstalling
		m.update.PhaseStatus[PhaseInstalling] = "running"
		return m, m.runInstallPhase()

	case UpdateInstallDoneMsg:
		// Install completed - transition to verify phase
		m.update.PhaseStatus[PhaseInstalling] = "done"
		m.update.Phase = PhaseVerifying
		m.update.PhaseStatus[PhaseVerifying] = "running"
		return m, m.runVerifyPhase(msg)

	case ChangelogLoadedMsg:
		if msg.Err != nil {
			m.update.Changelog = "Failed to load changelog: " + msg.Err.Error()
		} else {
			m.update.Changelog = msg.Content
		}
		m.clearChangelogModal() // Force rebuild with new content
		return m, nil

	case FocusPluginByIDMsg:
		// Switch to requested plugin
		return m, m.FocusPluginByID(msg.PluginID)

	case SwitchWorktreeMsg:
		// Switch to the requested worktree
		return m, m.switchWorktree(msg.WorktreePath)

	case WorktreeDeletedMsg:
		// Current worktree was deleted (detected by periodic check) - switch to main
		return m, tea.Batch(
			m.switchWorktree(msg.MainPath),
			ShowToast("Worktree deleted, switched to main", 3*time.Second),
		)

	case SwitchToMainWorktreeMsg:
		// Current worktree was deleted (detected by workspace plugin) - switch to main
		if msg.MainWorktreePath != "" && msg.MainWorktreePath != m.ui.WorkDir {
			return m, tea.Batch(
				m.switchProject(msg.MainWorktreePath),
				func() tea.Msg {
					return ToastMsg{
						Message:  "Worktree deleted, switched to main repo",
						Duration: 3 * time.Second,
					}
				},
			)
		}
		return m, nil

	case plugin.OpenFileMsg:
		// Open file in editor using tea.ExecProcess
		// Most editors support +lineNo syntax for opening at a line
		args := []string{}
		if msg.LineNo > 0 {
			args = append(args, fmt.Sprintf("+%d", msg.LineNo))
		}
		args = append(args, msg.Path)
		c := exec.Command(msg.Editor, args...)
		termState, _ := term.GetState(int(os.Stdout.Fd()))
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			if termState != nil {
				_ = term.Restore(int(os.Stdout.Fd()), termState)
			}
			return EditorReturnedMsg{Err: err}
		})

	case EditorReturnedMsg:
		// After editor exits, re-enable mouse and trigger refresh
		// tea.ExecProcess disables mouse, need to restore it
		cmds := []tea.Cmd{
			func() tea.Msg { return tea.EnableMouseAllMotion() },
		}
		if msg.Err != nil {
			cmds = append(cmds, func() tea.Msg { return ErrorMsg(msg) })
		} else {
			cmds = append(cmds, func() tea.Msg { return RefreshMsg{} })
		}
		return m, tea.Batch(cmds...)

	case palette.CommandSelectedMsg:
		// Execute the selected command from the palette
		m.showPalette = false
		m.updateContext()
		// Look up and execute the command
		if cmd, ok := m.keymap.GetCommand(msg.CommandID); ok && cmd.Handler != nil {
			return m, cmd.Handler()
		}
		return m, nil

	case version.UpdateAvailableMsg:
		m.updateAvailable = &msg
		m.update.InstallMethod = msg.InstallMethod
		m.clearDiagnosticsModal() // Force rebuild so modal picks up new update state
		m.ShowToast(
			fmt.Sprintf("Update %s available! Press ! for details", msg.LatestVersion),
			15*time.Second,
		)
		return m, nil

	case version.TdVersionMsg:
		m.tdVersionInfo = &msg
		m.clearDiagnosticsModal() // Force rebuild so modal picks up new version state
		// Show toast if td has an update (only if sidecar doesn't also have one)
		if msg.HasUpdate && m.updateAvailable == nil {
			m.ShowToast(
				fmt.Sprintf("td update %s available! Press ! for details", msg.LatestVersion),
				15*time.Second,
			)
		}
		return m, nil

	case IssuePreviewResultMsg:
		m.issue.PreviewLoading = false
		if msg.Error != nil {
			m.issue.PreviewError = msg.Error
		} else {
			m.issue.PreviewData = msg.Data
		}
		// Clear modal cache to trigger rebuild
		m.issue.PreviewModal = nil
		m.issue.PreviewModalWidth = 0
		return m, nil

	case IssueSearchResultMsg:
		// Discard stale results
		if msg.Query != m.issue.SearchQuery || !m.issue.ShowInput {
			return m, nil
		}
		m.issue.SearchLoading = false
		if msg.Error == nil {
			m.issue.SearchResults = msg.Results
		}
		m.issue.SearchScrollOffset = 0
		m.issue.InputModal = nil
		m.issue.InputModalWidth = 0
		return m, nil

	case ConfigChangedMsg:
		return m.handleConfigChanged(msg)
	}

	// Forward other messages to ALL plugins (not just active)
	// This ensures plugin-specific messages (like SessionsLoadedMsg) reach
	// their target plugin even when another plugin is focused
	plugins := m.registry.Plugins()
	for i, p := range plugins {
		newPlugin, cmd := p.Update(msg)
		plugins[i] = newPlugin
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if !m.showHelp && !m.showDiagnostics {
		m.updateContext()
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg processes keyboard input.
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Close modals with escape (priority order via activeModal)
	if msg.Type == tea.KeyEsc {
		switch m.activeModal() {
		case ModalPalette:
			m.showPalette = false
			m.updateContext()
			return m, nil
		case ModalHelp:
			m.showHelp = false
			m.clearHelpModal()
			return m, nil
		case ModalUpdate:
			// Handle Esc in update modal
			if m.update.ChangelogVisible {
				// Close changelog overlay, return to preview
				m.update.ChangelogVisible = false
				m.update.ChangelogScrollOffset = 0
				m.clearChangelogModal()
				return m, nil
			}
			// Close update modal
			m.update.ModalState = UpdateModalClosed
			return m, nil
		case ModalDiagnostics:
			m.showDiagnostics = false
			return m, nil
		case ModalQuitConfirm:
			m.showQuitConfirm = false
			return m, nil
		case ModalProjectSwitcher:
			// If in add mode, Esc exits back to list
			if m.project.AddMode {
				m.resetProjectAdd()
				return m, nil
			}
			// Esc: clear filter if set, otherwise close
			if m.project.Input.Value() != "" {
				m.project.Input.SetValue("")
				m.project.Filtered = m.cfg.Projects.List
				m.project.Cursor = 0
				m.project.Scroll = 0
				return m, nil
			}
			m.resetProjectSwitcher()
			m.updateContext()
			return m, nil
		case ModalWorktreeSwitcher:
			// Esc: clear filter if set, otherwise close
			if m.worktree.Input.Value() != "" {
				m.worktree.Input.SetValue("")
				m.worktree.Filtered = m.worktree.All
				m.worktree.Cursor = 0
				m.worktree.Scroll = 0
				return m, nil
			}
			m.resetWorktreeSwitcher()
			m.updateContext()
			return m, nil
		case ModalIssueInput:
			m.resetIssueInput()
			m.updateContext()
			return m, nil
		case ModalIssuePreview:
			m.resetIssuePreview()
			m.resetIssueInput()
			m.updateContext()
			return m, nil
		case ModalThemeSwitcher:
			// Esc: clear filter if set, otherwise close (restore original)
			if m.theme.Input.Value() != "" {
				m.theme.Input.SetValue("")
				m.theme.Filtered = buildUnifiedThemeList()
				m.theme.SelectedIdx = 0
				return m, nil
			}
			m.previewThemeEntry(m.theme.Original)
			m.resetThemeSwitcher()
			m.updateContext()
			return m, nil
		}
	}

	if m.showQuitConfirm {
		action, cmd := m.quitModal.HandleKey(msg)
		switch action {
		case "quit":
			// Save active plugin before quitting
			if activePlugin := m.ActivePlugin(); activePlugin != nil {
				_ = state.SetActivePlugin(m.ui.ProjectRoot, activePlugin.ID())
			}
			m.registry.Stop()
			return m, tea.Quit
		case "cancel":
			m.showQuitConfirm = false
			return m, nil
		}
		return m, cmd
	}

	// Handle update modal keys
	if m.update.ModalState != UpdateModalClosed {
		return m.handleUpdateModalKey(msg)
	}

	// Interactive/inline edit mode: forward ALL keys to plugin including ctrl+c
	// This ensures characters like `, ~, ?, !, @, q, 1-5 reach tmux instead of triggering app shortcuts
	// Ctrl+C is forwarded to tmux (to interrupt running processes) instead of showing quit dialog
	// User can exit interactive mode with Ctrl+\ first, then quit normally
	if m.activeContext == "workspace-interactive" || m.activeContext == "file-browser-inline-edit" || m.activeContext == "notes-inline-edit" {
		// Forward ALL keys to plugin (exit keys and ctrl+c handled by plugin)
		if p := m.ActivePlugin(); p != nil {
			newPlugin, cmd := p.Update(msg)
			plugins := m.registry.Plugins()
			if m.activePlugin < len(plugins) {
				plugins[m.activePlugin] = newPlugin
			}
			m.updateContext()
			return m, cmd
		}
		return m, nil
	}

	// Text input contexts: forward all keys to plugin except ctrl+c.
	// Uses plugin runtime capability first, then app-level fallback contexts.
	if m.consumesTextInput() {
		// ctrl+c shows quit confirmation
		if msg.String() == "ctrl+c" {
			if !m.hasModal() {
				m.initQuitModal()
				m.showQuitConfirm = true
			}
			return m, nil
		}
		// Forward everything else to plugin (esc, alt+enter handled by plugin)
		if p := m.ActivePlugin(); p != nil {
			newPlugin, cmd := p.Update(msg)
			plugins := m.registry.Plugins()
			if m.activePlugin < len(plugins) {
				plugins[m.activePlugin] = newPlugin
			}
			m.updateContext()
			return m, cmd
		}
		return m, nil
	}

	// Global quit - ctrl+c always takes precedence, 'q' in root plugin contexts
	switch msg.String() {
	case "ctrl+c":
		if !m.hasModal() {
			m.initQuitModal()
			m.showQuitConfirm = true
			return m, nil
		}
	case "q":
		if !m.hasModal() && isRootContext(m.activeContext) {
			m.initQuitModal()
			m.showQuitConfirm = true
			return m, nil
		}
		// Fall through to forward to plugin for navigation (back/escape)
	}

	// Handle palette input when open (Esc handled above)
	if m.showPalette {
		var cmd tea.Cmd
		m.palette, cmd = m.palette.Update(msg)
		return m, cmd
	}

	// Handle diagnostics modal keys
	if m.showDiagnostics {
		m.ensureDiagnosticsModal()
		if m.diagnosticsModal != nil {
			action, cmd := m.diagnosticsModal.HandleKey(msg)
			if cmd != nil {
				return m, cmd
			}
			switch action {
			case "update":
				// Open update modal instead of starting update directly
				if m.hasUpdatesAvailable() && !m.update.InProgress && !m.update.NeedsRestart {
					m.update.ReleaseNotes = ""
					if m.updateAvailable != nil {
						m.update.ReleaseNotes = m.updateAvailable.ReleaseNotes
					}
					m.update.ModalState = UpdateModalPreview
					m.showDiagnostics = false
					return m, nil
				}
			}
		}
		// Handle 'u' shortcut for update - open update modal
		if msg.String() == "u" && m.hasUpdatesAvailable() && !m.update.InProgress && !m.update.NeedsRestart {
			m.update.ReleaseNotes = ""
			if m.updateAvailable != nil {
				m.update.ReleaseNotes = m.updateAvailable.ReleaseNotes
			}
			m.update.ModalState = UpdateModalPreview
			m.showDiagnostics = false
			return m, nil
		}
		return m, nil
	}

	// Handle worktree switcher modal keys (Esc handled above)
	if m.worktree.Show {
		return m.handleWorktreeSwitcherKeys(msg)
	}

	// Handle project switcher modal keys (Esc handled above)
	if m.project.Show {
		return m.handleProjectSwitcherKeys(msg)
	}

	// Handle theme switcher modal keys (Esc handled above)
	if m.theme.Show {
		return m.handleThemeSwitcherKeys(msg)
	}

	// Handle issue input modal keys
	if m.issue.ShowInput {
		return m.handleIssueInputKeys(msg)
	}

	// Handle issue preview modal keys
	if m.issue.ShowPreview {
		return m.handleIssuePreviewKeys(msg)
	}

	// If any modal is open, don't process plugin/toggle keys
	if m.hasModal() {
		return m, nil
	}

	// Plugin switching (backtick, tilde, number keys)
	if mdl, cmd, handled := m.handlePluginSwitchKeys(msg); handled {
		return mdl, cmd
	}

	// Toggles (?, !, @, W, #, i, r)
	if mdl, cmd, handled := m.handleToggleKeys(msg); handled {
		return mdl, cmd
	}

	// Keymap routing + plugin forwarding
	return m.handleKeymapAndPluginForward(msg)
}

// updateContext sets activeContext based on current state.
func (m *Model) updateContext() {
	if p := m.ActivePlugin(); p != nil {
		m.activeContext = p.FocusContext().String()
	} else {
		m.activeContext = "global"
	}
}

// consumesTextInput returns true when the active context should treat printable
// keys as text input (block app-level navigation shortcuts).
func (m *Model) consumesTextInput() bool {
	if p := m.ActivePlugin(); p != nil {
		if c, ok := p.(plugin.TextInputConsumer); ok && c.ConsumesTextInput() {
			return true
		}
	}
	return isTextInputContext(m.activeContext)
}

// isRootContext returns true if the context is a root view where 'q' should quit.
// Root contexts are plugin top-level views (not sub-views like detail/diff/commit).
func isRootContext(ctx string) bool {
	fc := keymap.FocusContext(ctx)
	return fc.IsRoot()
}

// isTextInputContext returns true if the context is a text input mode
// where alphanumeric keys should be forwarded to the plugin for typing.
func isTextInputContext(ctx string) bool {
	switch ctx {
	case "td-search", "td-form", "td-board-editor", "td-confirm", "td-close-confirm",
		"theme-switcher",
		"issue-input":
		return true
	default:
		return false
	}
}

// isGlobalRefreshContext returns true if 'r' should trigger a global refresh.
// Returns false for contexts where 'r' should be forwarded to the plugin
// (text input modes or plugin-specific 'r' bindings).
func isGlobalRefreshContext(ctx string) bool {
	switch ctx {
	// Global context - 'r' refreshes
	case "global", "":
		return true

	// Git status contexts - 'r' refreshes (no text input, no 'r' binding)
	case "git-status", "git-history", "git-commit-detail", "git-diff":
		return true

	// Conversations list - 'r' refreshes (no text input, no 'r' binding)
	case "conversations", "conversation-detail", "message-detail":
		return true

	// File browser preview - 'r' refreshes (no text input)
	case "file-browser-preview":
		return true

	// Contexts where 'r' should be forwarded to plugin:
	// - td-monitor: 'r' is mark-review
	// - file-browser-tree: 'r' is rename
	// - file-browser-search: text input mode
	// - file-browser-content-search: text input mode
	// - file-browser-quick-open: text input mode
	// - file-browser-file-op: text input mode
	// - conversations-search: text input mode
	// - conversations-filter: text input mode
	// - git-commit: text input mode (commit message)
	// - td-modal: modal view
	// - palette: command palette
	// - diagnostics: diagnostics view
	default:
		return false
	}
}

// handleQuitConfirmMouse handles mouse events for the quit confirmation modal.
func (m *Model) handleHelpModalMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureHelpModal()
	if m.helpModal == nil {
		return m, nil
	}
	// Info-only modal - no mouse interaction needed beyond ensuring modal exists
	return m, nil
}

// handleUpdateModalKey handles keyboard input for the update modal.
func (m *Model) handleUpdateModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle changelog overlay first if visible
	if m.update.ChangelogVisible {
		switch key {
		case "j", "down":
			m.update.ChangelogScrollOffset++
			m.syncChangelogScroll()
			return m, nil
		case "k", "up":
			if m.update.ChangelogScrollOffset > 0 {
				m.update.ChangelogScrollOffset--
				m.syncChangelogScroll()
			}
			return m, nil
		case "ctrl+d", "pgdown":
			m.update.ChangelogScrollOffset += 10
			m.syncChangelogScroll()
			return m, nil
		case "ctrl+u", "pgup":
			m.update.ChangelogScrollOffset -= 10
			if m.update.ChangelogScrollOffset < 0 {
				m.update.ChangelogScrollOffset = 0
			}
			m.syncChangelogScroll()
			return m, nil
		case "g":
			m.update.ChangelogScrollOffset = 0
			m.syncChangelogScroll()
			return m, nil
		case "G":
			m.update.ChangelogScrollOffset = 999999 // Will be clamped during render
			m.syncChangelogScroll()
			return m, nil
		case "esc", "c", "q":
			m.update.ChangelogVisible = false
			m.update.ChangelogScrollOffset = 0
			m.clearChangelogModal()
			return m, nil
		}
		// Route to modal for Enter (close button)
		m.ensureChangelogModal()
		if m.update.ChangelogModal != nil {
			action, _ := m.update.ChangelogModal.HandleKey(msg)
			if action == "cancel" {
				m.update.ChangelogVisible = false
				m.update.ChangelogScrollOffset = 0
				m.clearChangelogModal()
				return m, nil
			}
		}
		return m, nil
	}

	// Handle keys based on modal state
	switch m.update.ModalState {
	case UpdateModalPreview:
		// Handle special keys first
		switch key {
		case "c":
			// Show changelog
			m.update.ChangelogScrollOffset = 0
			if m.update.Changelog == "" {
				m.update.ChangelogVisible = true
				return m, fetchChangelog()
			}
			m.update.ChangelogVisible = true
			return m, nil
		case "q":
			m.update.ModalState = UpdateModalClosed
			return m, nil
		}
		// Route to modal for Tab/Shift+Tab/Enter/Esc
		m.ensureUpdatePreviewModal()
		if m.update.PreviewModal != nil {
			action, cmd := m.update.PreviewModal.HandleKey(msg)
			switch action {
			case "update":
				m.update.ModalState = UpdateModalProgress
				m.update.InProgress = true
				m.update.StartTime = time.Now()
				m.initPhaseStatus()
				m.update.Phase = PhaseCheckPrereqs
				m.update.PhaseStatus[PhaseCheckPrereqs] = "running"
				return m, m.startUpdateWithPhases()
			case "cancel":
				m.update.ModalState = UpdateModalClosed
				return m, nil
			}
			if cmd != nil {
				return m, cmd
			}
		}

	case UpdateModalProgress:
		// No keys during progress (except Esc handled earlier)
		return m, nil

	case UpdateModalComplete:
		// Handle 'q' specially for quit
		if key == "q" {
			if activePlugin := m.ActivePlugin(); activePlugin != nil {
				_ = state.SetActivePlugin(m.ui.ProjectRoot, activePlugin.ID())
			}
			m.registry.Stop()
			return m, tea.Quit
		}
		// Route to modal for Tab/Shift+Tab/Enter/Esc
		m.ensureUpdateCompleteModal()
		if m.update.CompleteModal != nil {
			action, cmd := m.update.CompleteModal.HandleKey(msg)
			switch action {
			case "quit":
				if activePlugin := m.ActivePlugin(); activePlugin != nil {
					_ = state.SetActivePlugin(m.ui.ProjectRoot, activePlugin.ID())
				}
				m.registry.Stop()
				return m, tea.Quit
			case "cancel":
				m.update.ModalState = UpdateModalClosed
				return m, nil
			}
			if cmd != nil {
				return m, cmd
			}
		}

	case UpdateModalError:
		// Handle 'r' for retry and 'q' for close
		switch key {
		case "r":
			m.update.ModalState = UpdateModalProgress
			m.update.Error = ""
			m.update.StartTime = time.Now()
			m.initPhaseStatus()
			m.update.Phase = PhaseCheckPrereqs
			m.update.PhaseStatus[PhaseCheckPrereqs] = "running"
			return m, m.startUpdateWithPhases()
		case "q":
			m.update.ModalState = UpdateModalClosed
			return m, nil
		}
		// Route to modal for Tab/Shift+Tab/Enter/Esc
		m.ensureUpdateErrorModal()
		if m.update.ErrorModal != nil {
			action, cmd := m.update.ErrorModal.HandleKey(msg)
			switch action {
			case "retry":
				m.update.ModalState = UpdateModalProgress
				m.update.Error = ""
				m.update.StartTime = time.Now()
				m.initPhaseStatus()
				m.update.Phase = PhaseCheckPrereqs
				m.update.PhaseStatus[PhaseCheckPrereqs] = "running"
				return m, m.startUpdateWithPhases()
			case "cancel":
				m.update.ModalState = UpdateModalClosed
				return m, nil
			}
			if cmd != nil {
				return m, cmd
			}
		}
	}

	return m, nil
}

// handleUpdateModalMouse handles mouse events for the update modal.
func (m *Model) handleUpdateModalMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Handle changelog overlay first if visible
	if m.update.ChangelogVisible {
		m.ensureChangelogModal()
		if m.update.ChangelogMouseHandler == nil {
			m.update.ChangelogMouseHandler = mouse.NewHandler()
		}
		// Handle scroll events via shared state pointer (no modal rebuild needed)
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if m.update.ChangelogScrollOffset > 0 {
				m.update.ChangelogScrollOffset -= 3
				if m.update.ChangelogScrollOffset < 0 {
					m.update.ChangelogScrollOffset = 0
				}
				m.syncChangelogScroll()
			}
			return m, nil
		case tea.MouseButtonWheelDown:
			m.update.ChangelogScrollOffset += 3
			m.syncChangelogScroll()
			return m, nil
		}
		// Handle modal interaction (close button, backdrop)
		if m.update.ChangelogModal != nil {
			action := m.update.ChangelogModal.HandleMouse(msg, m.update.ChangelogMouseHandler)
			if action == "cancel" {
				m.update.ChangelogVisible = false
				m.update.ChangelogScrollOffset = 0
				m.clearChangelogModal()
				return m, nil
			}
		}
		return m, nil
	}

	switch m.update.ModalState {
	case UpdateModalPreview:
		m.ensureUpdatePreviewModal()
		if m.update.PreviewModal == nil {
			return m, nil
		}
		if m.update.PreviewMouseHandler == nil {
			m.update.PreviewMouseHandler = mouse.NewHandler()
		}
		action := m.update.PreviewModal.HandleMouse(msg, m.update.PreviewMouseHandler)
		switch action {
		case "update":
			m.update.ModalState = UpdateModalProgress
			m.update.InProgress = true
			m.update.StartTime = time.Now()
			m.initPhaseStatus()
			m.update.Phase = PhaseCheckPrereqs
			m.update.PhaseStatus[PhaseCheckPrereqs] = "running"
			return m, m.startUpdateWithPhases()
		case "cancel":
			m.update.ModalState = UpdateModalClosed
			return m, nil
		}

	case UpdateModalComplete:
		m.ensureUpdateCompleteModal()
		if m.update.CompleteModal == nil {
			return m, nil
		}
		if m.update.CompleteMouseHandler == nil {
			m.update.CompleteMouseHandler = mouse.NewHandler()
		}
		action := m.update.CompleteModal.HandleMouse(msg, m.update.CompleteMouseHandler)
		switch action {
		case "quit":
			if activePlugin := m.ActivePlugin(); activePlugin != nil {
				_ = state.SetActivePlugin(m.ui.ProjectRoot, activePlugin.ID())
			}
			m.registry.Stop()
			return m, tea.Quit
		case "cancel":
			m.update.ModalState = UpdateModalClosed
			return m, nil
		}

	case UpdateModalError:
		m.ensureUpdateErrorModal()
		if m.update.ErrorModal == nil {
			return m, nil
		}
		if m.update.ErrorMouseHandler == nil {
			m.update.ErrorMouseHandler = mouse.NewHandler()
		}
		action := m.update.ErrorModal.HandleMouse(msg, m.update.ErrorMouseHandler)
		switch action {
		case "retry":
			m.update.ModalState = UpdateModalProgress
			m.update.Error = ""
			m.update.StartTime = time.Now()
			m.initPhaseStatus()
			m.update.Phase = PhaseCheckPrereqs
			m.update.PhaseStatus[PhaseCheckPrereqs] = "running"
			return m, m.startUpdateWithPhases()
		case "cancel":
			m.update.ModalState = UpdateModalClosed
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) handleQuitConfirmMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	action := m.quitModal.HandleMouse(msg, m.quitMouseHandler)
	switch action {
	case "quit":
		// Save active plugin before quitting
		if activePlugin := m.ActivePlugin(); activePlugin != nil {
			_ = state.SetActivePlugin(m.ui.ProjectRoot, activePlugin.ID())
		}
		m.registry.Stop()
		return m, tea.Quit
	case "cancel":
		m.showQuitConfirm = false
		return m, nil
	}
	return m, nil
}
