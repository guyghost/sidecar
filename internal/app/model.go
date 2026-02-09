package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/community"
	"github.com/guyghost/sidecar/internal/config"
	"github.com/guyghost/sidecar/internal/keymap"
	"github.com/guyghost/sidecar/internal/modal"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/palette"
	"github.com/guyghost/sidecar/internal/plugin"
	"github.com/guyghost/sidecar/internal/state"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/theme"
	"github.com/guyghost/sidecar/internal/version"
)

// ModalKind identifies an app-level modal with explicit priority ordering.
// Lower values = higher priority (checked first for rendering and input routing).
type ModalKind int

const (
	ModalNone             ModalKind = iota // No modal open
	ModalPalette                           // Command palette (highest priority)
	ModalHelp                              // Help overlay
	ModalUpdate                            // Update modal
	ModalDiagnostics                       // Diagnostics/version info
	ModalQuitConfirm                       // Quit confirmation dialog
	ModalProjectSwitcher                   // Project switcher
	ModalWorktreeSwitcher                  // Worktree switcher
	ModalThemeSwitcher                     // Theme switcher
	ModalIssueInput                        // Issue ID text input
	ModalIssuePreview                      // Issue preview display (lowest priority)
)

// activeModal returns the highest-priority open modal.
// This is the single source of truth for which modal is currently active.
func (m *Model) activeModal() ModalKind {
	switch {
	case m.showPalette:
		return ModalPalette
	case m.showHelp:
		return ModalHelp
	case m.update.ModalState != UpdateModalClosed:
		return ModalUpdate
	case m.showDiagnostics:
		return ModalDiagnostics
	case m.showQuitConfirm:
		return ModalQuitConfirm
	case m.project.Show:
		return ModalProjectSwitcher
	case m.worktree.Show:
		return ModalWorktreeSwitcher
	case m.theme.Show:
		return ModalThemeSwitcher
	case m.issue.ShowInput:
		return ModalIssueInput
	case m.issue.ShowPreview:
		return ModalIssuePreview
	default:
		return ModalNone
	}
}

// hasModal returns true if any app-level modal is open.
func (m *Model) hasModal() bool {
	return m.activeModal() != ModalNone
}

// TabBounds represents the X position range of a tab for mouse hit testing.
type TabBounds struct {
	Start, End int
}

type projectAddState struct {
	nameInput     textinput.Model
	pathInput     textinput.Model
	errorMessage  string
	themeSelected string
}

// ProjectSwitcherState holds all state for the project switcher modal.
type ProjectSwitcherState struct {
	Show         bool
	Cursor       int
	Scroll       int
	Input        textinput.Model
	Filtered     []config.ProjectConfig
	Modal        *modal.Modal
	ModalWidth   int
	MouseHandler *mouse.Handler
	// Add sub-mode (within project switcher)
	AddMode         bool
	Add             *projectAddState
	AddModal        *modal.Modal
	AddModalWidth   int
	AddMouseHandler *mouse.Handler
	// Theme picker within add-project flow
	AddThemeMode       bool
	AddThemeCursor     int
	AddThemeScroll     int
	AddThemeInput      textinput.Model
	AddThemeFiltered   []string // filtered built-in theme list
	AddCommunityMode   bool     // in community sub-browser?
	AddCommunityList   []string // filtered community scheme names
	AddCommunityCursor int
	AddCommunityScroll int
}

// WorktreeSwitcherState holds all state for the worktree switcher modal.
type WorktreeSwitcherState struct {
	Show         bool
	Cursor       int
	Scroll       int
	Input        textinput.Model
	Filtered     []WorktreeInfo
	All          []WorktreeInfo
	Modal        *modal.Modal
	ModalWidth   int
	MouseHandler *mouse.Handler
	CheckCounter int           // Counter for periodic worktree existence check
	CachedInfo   *WorktreeInfo // avoids git subprocess forks on every View render
}

// ThemeSwitcherState holds all state for the theme switcher modal.
type ThemeSwitcherState struct {
	Show         bool
	Modal        *modal.Modal
	ModalWidth   int
	MouseHandler *mouse.Handler
	SelectedIdx  int
	Input        textinput.Model
	Filtered     []themeEntry
	Original     themeEntry // original theme to restore on cancel
	Scope        string     // "global" or "project"
}

// IssueState holds all state for the issue input and preview modals.
type IssueState struct {
	// Input phase
	ShowInput         bool
	InputModel        textinput.Model
	InputModal        *modal.Modal
	InputModalWidth   int
	InputMouseHandler *mouse.Handler
	// Search auto-complete
	SearchResults       []IssueSearchResult
	SearchQuery         string // last query sent to td search
	SearchLoading       bool
	SearchCursor        int  // selected result index (-1 = none/input focused)
	SearchScrollOffset  int  // viewport scroll offset for search results
	SearchIncludeClosed bool // whether to include closed issues in search
	// Preview phase
	ShowPreview         bool
	PreviewData         *IssuePreviewData
	PreviewLoading      bool
	PreviewError        error
	PreviewModal        *modal.Modal
	PreviewModalWidth   int
	PreviewMouseHandler *mouse.Handler
}

// UpdateState holds all state for the update modal and update process.
type UpdateState struct {
	InProgress    bool
	Error         string
	NeedsRestart  bool
	SpinnerFrame  int
	InstallMethod version.InstallMethod
	// Modal state
	ModalState            UpdateModalState
	Phase                 UpdatePhase
	PhaseStatus           map[UpdatePhase]string
	StartTime             time.Time
	ReleaseNotes          string // Release notes for current update
	Changelog             string // Full changelog content
	ChangelogVisible      bool
	ChangelogScrollOffset int
	ChangelogScrollState  *changelogViewState // Shared state for modal closure
	// Declarative modals
	PreviewModal             *modal.Modal
	PreviewModalWidth        int
	PreviewMouseHandler      *mouse.Handler
	CompleteModal            *modal.Modal
	CompleteModalWidth       int
	CompleteMouseHandler     *mouse.Handler
	ErrorModal               *modal.Modal
	ErrorModalWidth          int
	ErrorMouseHandler        *mouse.Handler
	ChangelogModal           *modal.Modal
	ChangelogModalWidth      int
	ChangelogMouseHandler    *mouse.Handler
	ChangelogRenderedLines   []string // Cached rendered changelog lines
	ChangelogMaxVisibleLines int      // Max lines visible in viewport
}

// Model is the root Bubble Tea model for the sidecar application.
type Model struct {
	// Configuration
	cfg *config.Config

	// Plugin management
	registry     *plugin.Registry
	activePlugin int

	// Keymap
	keymap        *keymap.Registry
	activeContext string

	// UI state
	width, height           int
	showHelp                bool
	helpModal               *modal.Modal
	helpModalWidth          int
	helpMouseHandler        *mouse.Handler
	showDiagnostics         bool
	diagnosticsModal        *modal.Modal
	diagnosticsModalWidth   int
	diagnosticsMouseHandler *mouse.Handler
	showClock               bool
	showPalette             bool
	showQuitConfirm         bool
	quitModal               *modal.Modal
	quitMouseHandler        *mouse.Handler
	palette                 palette.Model

	// Domain-specific modal state (extracted sub-structs)
	project  ProjectSwitcherState
	worktree WorktreeSwitcherState
	theme    ThemeSwitcherState
	issue    IssueState
	update   UpdateState

	// Header/footer
	ui *UIState

	// Status/toast messages
	statusMsg     string
	statusExpiry  time.Time
	statusIsError bool

	// Error handling
	lastError error

	// Ready state
	ready bool

	// Version info
	currentVersion  string
	updateAvailable *version.UpdateAvailableMsg
	tdVersionInfo   *version.TdVersionMsg

	// Intro animation
	intro IntroModel
}

// New creates a new application model.
// initialPluginID optionally specifies which plugin to focus on startup (empty = first plugin).
func New(reg *plugin.Registry, km *keymap.Registry, cfg *config.Config, currentVersion, workDir, projectRoot, initialPluginID string) Model {
	repoName := GetRepoName(workDir)
	ui := NewUIState()
	ui.WorkDir = workDir
	ui.ProjectRoot = projectRoot

	// Determine initial active plugin index
	activeIdx := 0
	if initialPluginID != "" {
		for i, p := range reg.Plugins() {
			if p.ID() == initialPluginID {
				activeIdx = i
				break
			}
		}
	}

	return Model{
		cfg:            cfg,
		registry:       reg,
		keymap:         km,
		activePlugin:   activeIdx,
		activeContext:  "global",
		showClock:      cfg.UI.ShowClock,
		palette:        palette.New(),
		ui:             ui,
		ready:          false,
		intro:          NewIntroModel(repoName),
		currentVersion: currentVersion,
		update:         UpdateState{PhaseStatus: make(map[UpdatePhase]string)},
	}
}

// Init initializes the model and returns initial commands.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tickCmd(),
		IntroTick(),
		version.CheckAsync(m.currentVersion),
		version.CheckTdAsync(),
	}

	// Start all registered plugins
	for _, cmd := range m.registry.Start() {
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// initQuitModal initializes the quit confirmation modal.
func (m *Model) initQuitModal() {
	m.quitModal = modal.New("Quit Sidecar?",
		modal.WithWidth(50),
		modal.WithVariant(modal.VariantDefault),
		modal.WithPrimaryAction("quit"),
	).
		AddSection(modal.Text("Are you sure you want to quit?")).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(" Quit ", "quit"),
			modal.Btn(" Cancel ", "cancel"),
		))
	m.quitMouseHandler = mouse.NewHandler()
}

// ActivePlugin returns the currently active plugin.
func (m Model) ActivePlugin() plugin.Plugin {
	plugins := m.registry.Plugins()
	if len(plugins) == 0 {
		return nil
	}
	if m.activePlugin >= len(plugins) {
		m.activePlugin = 0
	}
	return plugins[m.activePlugin]
}

// SetActivePlugin sets the active plugin by index and returns a command
// to notify the plugin it has been focused.
func (m *Model) SetActivePlugin(idx int) tea.Cmd {
	plugins := m.registry.Plugins()
	if idx >= 0 && idx < len(plugins) {
		// Unfocus current
		if current := m.ActivePlugin(); current != nil {
			current.SetFocused(false)
		}
		m.activePlugin = idx
		// Focus new
		if next := m.ActivePlugin(); next != nil {
			next.SetFocused(true)
			m.activeContext = next.FocusContext().String()
			return PluginFocused()
		}
	}
	return nil
}

// NextPlugin switches to the next plugin.
func (m *Model) NextPlugin() tea.Cmd {
	plugins := m.registry.Plugins()
	if len(plugins) == 0 {
		return nil
	}
	return m.SetActivePlugin((m.activePlugin + 1) % len(plugins))
}

// PrevPlugin switches to the previous plugin.
func (m *Model) PrevPlugin() tea.Cmd {
	plugins := m.registry.Plugins()
	if len(plugins) == 0 {
		return nil
	}
	idx := m.activePlugin - 1
	if idx < 0 {
		idx = len(plugins) - 1
	}
	return m.SetActivePlugin(idx)
}

// FocusPluginByID switches to a plugin by its ID.
func (m *Model) FocusPluginByID(id string) tea.Cmd {
	plugins := m.registry.Plugins()
	for i, p := range plugins {
		if p.ID() == id {
			return m.SetActivePlugin(i)
		}
	}
	return nil
}

// ShowToast displays a temporary status message.
func (m *Model) ShowToast(msg string, duration time.Duration) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(duration)
}

// ClearToast clears any expired toast message.
func (m *Model) ClearToast() {
	if m.statusMsg != "" && time.Now().After(m.statusExpiry) {
		m.statusMsg = ""
		m.statusIsError = false
	}
}

// hasUpdatesAvailable returns true if either sidecar or td has an update available.
func (m *Model) hasUpdatesAvailable() bool {
	if m.updateAvailable != nil {
		return true
	}
	if m.tdVersionInfo != nil && m.tdVersionInfo.HasUpdate && m.tdVersionInfo.Installed {
		return true
	}
	return false
}

// initPhaseStatus initializes all update phases to pending status.
func (m *Model) initPhaseStatus() {
	m.update.PhaseStatus = map[UpdatePhase]string{
		PhaseCheckPrereqs: "pending",
		PhaseInstalling:   "pending",
		PhaseVerifying:    "pending",
	}
}

// startUpdateWithPhases starts the update process with phase tracking.
// This replaces the old doUpdate for the new modal-based update flow.
func (m *Model) startUpdateWithPhases() tea.Cmd {
	return tea.Batch(
		// Start elapsed timer
		m.startElapsedTimer(),
		// Start spinner animation
		updateSpinnerTick(),
		// Start the first phase (check prerequisites)
		m.runCheckPrerequisites(),
	)
}

// startElapsedTimer starts the elapsed time ticker.
func (m *Model) startElapsedTimer() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return UpdateElapsedTickMsg{}
	})
}

// UpdatePrereqsPassedMsg signals prerequisites check passed, triggers install phase.
type UpdatePrereqsPassedMsg struct{}

// UpdateInstallDoneMsg signals install completed, triggers verify phase.
type UpdateInstallDoneMsg struct {
	SidecarUpdated    bool
	TdUpdated         bool
	NewSidecarVersion string
	NewTdVersion      string
}

// runCheckPrerequisites runs the prerequisites check phase.
func (m *Model) runCheckPrerequisites() tea.Cmd {
	method := m.update.InstallMethod
	return func() tea.Msg {
		switch method {
		case version.InstallMethodHomebrew:
			if _, err := exec.LookPath("brew"); err != nil {
				return UpdateErrorMsg{Step: "check", Err: fmt.Errorf("brew not found in PATH")}
			}
		case version.InstallMethodBinary:
			// No prerequisites for binary download — just show the URL
		default:
			if _, err := exec.LookPath("go"); err != nil {
				return UpdateErrorMsg{Step: "check", Err: fmt.Errorf("go not found in PATH")}
			}
		}
		return UpdatePrereqsPassedMsg{}
	}
}

// runInstallPhase runs the install phase using the detected install method.
func (m *Model) runInstallPhase() tea.Cmd {
	sidecarUpdate := m.updateAvailable
	tdUpdate := m.tdVersionInfo
	method := m.update.InstallMethod

	return func() tea.Msg {
		var sidecarUpdated, tdUpdated bool
		var newSidecarVersion, newTdVersion string

		// Update sidecar
		if sidecarUpdate != nil {
			switch method {
			case version.InstallMethodHomebrew:
				cmd := exec.Command("brew", "upgrade", "sidecar")
				if output, err := cmd.CombinedOutput(); err != nil {
					return UpdateErrorMsg{Step: "sidecar", Err: fmt.Errorf("%v: %s", err, output)}
				}
				sidecarUpdated = true
				newSidecarVersion = sidecarUpdate.LatestVersion
			case version.InstallMethodBinary:
				// Binary installs cannot be auto-updated; the preview modal
				// shows the download URL instead. The user must download manually.
			default: // Go install
				args := []string{
					"install",
					"-ldflags", fmt.Sprintf("-X main.Version=%s", sidecarUpdate.LatestVersion),
					fmt.Sprintf("github.com/guyghost/sidecar/cmd/sidecar@%s", sidecarUpdate.LatestVersion),
				}
				cmd := exec.Command("go", args...)
				if output, err := cmd.CombinedOutput(); err != nil {
					return UpdateErrorMsg{Step: "sidecar", Err: fmt.Errorf("%v: %s", err, output)}
				}
				sidecarUpdated = true
				newSidecarVersion = sidecarUpdate.LatestVersion
			}
		}

		// Update td
		if tdUpdate != nil && tdUpdate.HasUpdate && tdUpdate.Installed {
			switch method {
			case version.InstallMethodHomebrew:
				cmd := exec.Command("brew", "upgrade", "td")
				if output, err := cmd.CombinedOutput(); err != nil {
					return UpdateErrorMsg{Step: "td", Err: fmt.Errorf("%v: %s", err, output)}
				}
			default: // Go install (binary users of td still use go install)
				cmd := exec.Command("go", "install",
					fmt.Sprintf("github.com/marcus/td@%s", tdUpdate.LatestVersion))
				if output, err := cmd.CombinedOutput(); err != nil {
					return UpdateErrorMsg{Step: "td", Err: fmt.Errorf("%v: %s", err, output)}
				}
			}
			tdUpdated = true
			newTdVersion = tdUpdate.LatestVersion
		}

		return UpdateInstallDoneMsg{
			SidecarUpdated:    sidecarUpdated,
			TdUpdated:         tdUpdated,
			NewSidecarVersion: newSidecarVersion,
			NewTdVersion:      newTdVersion,
		}
	}
}

// runVerifyPhase runs the verification phase (check installed binaries).
func (m *Model) runVerifyPhase(installResult UpdateInstallDoneMsg) tea.Cmd {
	return func() tea.Msg {
		// Verify sidecar binary if it was updated
		if installResult.SidecarUpdated {
			sidecarPath, err := exec.LookPath("sidecar")
			if err != nil {
				return UpdateErrorMsg{Step: "verify", Err: fmt.Errorf("sidecar not found in PATH after install")}
			}
			// Verify the binary is executable by running --version
			cmd := exec.Command(sidecarPath, "--version")
			if err := cmd.Run(); err != nil {
				return UpdateErrorMsg{Step: "verify", Err: fmt.Errorf("sidecar binary not executable: %v", err)}
			}
		}

		// Verify td binary if it was updated
		if installResult.TdUpdated {
			tdPath, err := exec.LookPath("td")
			if err != nil {
				return UpdateErrorMsg{Step: "verify", Err: fmt.Errorf("td not found in PATH after install")}
			}
			// Verify the binary is executable
			cmd := exec.Command(tdPath, "version", "--short")
			if err := cmd.Run(); err != nil {
				return UpdateErrorMsg{Step: "verify", Err: fmt.Errorf("td binary not executable: %v", err)}
			}
		}

		return UpdateSuccessMsg(installResult)
	}
}

// resetProjectSwitcher resets the project switcher modal state.
func (m *Model) resetProjectSwitcher() {
	m.project.Show = false
	m.project.Cursor = 0
	m.project.Scroll = 0
	m.project.Filtered = nil
	m.clearProjectSwitcherModal()
	m.resetProjectAdd()
	// Restore current project's theme (undo any live preview)
	resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
	theme.ApplyResolved(resolved)
}

// clearProjectSwitcherModal clears the modal cache.
func (m *Model) clearProjectSwitcherModal() {
	m.project.Modal = nil
	m.project.ModalWidth = 0
	m.project.MouseHandler = nil
}

// initProjectSwitcher initializes the project switcher modal.
func (m *Model) initProjectSwitcher() {
	m.clearProjectSwitcherModal()
	ti := textinput.New()
	ti.Placeholder = "Filter projects..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	m.project.Input = ti
	m.project.Filtered = m.cfg.Projects.List
	m.project.Cursor = 0
	m.project.Scroll = 0

	// Set cursor to current project if found
	for i, proj := range m.project.Filtered {
		if proj.Path == m.ui.WorkDir {
			m.project.Cursor = i
			break
		}
	}
	// Preview the initially-selected project's theme
	m.previewProjectTheme()
}

// filterProjects filters projects by name or path using a case-insensitive substring match.
func filterProjects(all []config.ProjectConfig, query string) []config.ProjectConfig {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	var matches []config.ProjectConfig
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Path), q) {
			matches = append(matches, p)
		}
	}
	return matches
}

// projectSwitcherEnsureCursorVisible adjusts scroll to keep cursor in view.
// Returns the new scroll offset.
func projectSwitcherEnsureCursorVisible(cursor, scroll, maxVisible int) int {
	if cursor < scroll {
		return cursor
	}
	if cursor >= scroll+maxVisible {
		return cursor - maxVisible + 1
	}
	return scroll
}

// switchProject switches all plugins to a new project directory.
func (m *Model) switchProject(projectPath string) tea.Cmd {
	// Skip if already on this project
	if projectPath == m.ui.WorkDir {
		return func() tea.Msg {
			return ToastMsg{Message: "Already on this project", Duration: 2 * time.Second}
		}
	}

	// Save the active plugin state for the old project root
	oldWorkDir := m.ui.WorkDir
	oldProjectRoot := m.ui.ProjectRoot
	if activePlugin := m.ActivePlugin(); activePlugin != nil {
		_ = state.SetActivePlugin(oldProjectRoot, activePlugin.ID())
	}

	// Normalize old workdir for comparisons
	normalizedOldWorkDir, _ := normalizePath(oldWorkDir)

	// Check if target project has a saved worktree we should restore.
	// Only restore if projectPath is the main repo - if user explicitly chose a
	// specific worktree path (via worktree switcher), respect that choice.
	targetPath := projectPath
	if targetMainRepo := GetMainWorktreePath(projectPath); targetMainRepo != "" {
		normalizedProject, _ := normalizePath(projectPath)
		normalizedTargetMain, _ := normalizePath(targetMainRepo)

		// Only restore saved worktree if switching to the main repo path
		if normalizedProject == normalizedTargetMain {
			if savedWorktree := state.GetLastWorktreePath(normalizedTargetMain); savedWorktree != "" {
				// Don't restore if the saved worktree is where we're coming FROM
				// (user is explicitly leaving that worktree)
				if savedWorktree != normalizedOldWorkDir {
					// Verify saved worktree still exists
					if WorktreeExists(savedWorktree) {
						targetPath = savedWorktree
					} else {
						// Stale entry - clear it
						_ = state.ClearLastWorktreePath(normalizedTargetMain)
					}
				}
			}
		}

		// Save the final target as last active worktree for this repo
		normalizedTarget, _ := normalizePath(targetPath)
		_ = state.SetLastWorktreePath(normalizedTargetMain, normalizedTarget)
	}

	// Update the UI state
	m.ui.WorkDir = targetPath
	m.intro.RepoName = GetRepoName(targetPath)
	// Eagerly refresh worktree cache (must happen in Update, not View, due to value receiver)
	m.refreshWorktreeCache()

	// Resolve project root (main worktree for linked worktrees, same as targetPath otherwise)
	newProjectRoot := GetMainWorktreePath(targetPath)
	if newProjectRoot == "" {
		newProjectRoot = targetPath
	}
	m.ui.ProjectRoot = newProjectRoot

	// Apply project-specific theme (or global fallback)
	resolved := theme.ResolveTheme(m.cfg, targetPath)
	theme.ApplyResolved(resolved)

	// Reinitialize all plugins with the new working directory and project root
	// This stops all plugins, updates the context, and starts them again
	startCmds := m.registry.Reinit(targetPath, newProjectRoot)

	// Send WindowSizeMsg to all plugins so they recalculate layout/bounds.
	// Without this, plugins like td-monitor lose mouse interactivity because
	// their panel bounds are only calculated on WindowSizeMsg receipt.
	adjustedHeight := m.height - headerHeight - footerHeight
	sizeMsg := tea.WindowSizeMsg{Width: m.width, Height: adjustedHeight}
	plugins := m.registry.Plugins()
	for i, p := range plugins {
		newPlugin, cmd := p.Update(sizeMsg)
		plugins[i] = newPlugin
		if cmd != nil {
			startCmds = append(startCmds, cmd)
		}
	}

	// Restore active plugin for the new project root if saved, otherwise keep current
	newActivePluginID := state.GetActivePlugin(newProjectRoot)
	if newActivePluginID != "" {
		m.FocusPluginByID(newActivePluginID)
	}

	// Return batch of start commands plus a toast notification
	return tea.Batch(
		tea.Batch(startCmds...),
		func() tea.Msg {
			return ToastMsg{
				Message:  fmt.Sprintf("Switched to %s", GetRepoName(targetPath)),
				Duration: 3 * time.Second,
			}
		},
	)
}

// previewProjectTheme applies the theme for the currently selected project in the switcher.
func (m *Model) previewProjectTheme() {
	projects := m.project.Filtered
	if m.project.Cursor >= 0 && m.project.Cursor < len(projects) {
		resolved := theme.ResolveTheme(m.cfg, projects[m.project.Cursor].Path)
		theme.ApplyResolved(resolved)
	}
}

// currentProjectConfig returns the ProjectConfig for the current workdir, or nil.
// If the current workdir is a worktree, it also checks if the main worktree path
// matches a registered project (so theme scope selector works from worktrees).
func (m *Model) currentProjectConfig() *config.ProjectConfig {
	if m.cfg == nil {
		return nil
	}
	// First, check direct match
	for i := range m.cfg.Projects.List {
		if m.cfg.Projects.List[i].Path == m.ui.WorkDir {
			return &m.cfg.Projects.List[i]
		}
	}

	// If not found, check if we're in a worktree and the main repo is registered
	mainPath := GetMainWorktreePath(m.ui.WorkDir)
	if mainPath != "" && mainPath != m.ui.WorkDir {
		for i := range m.cfg.Projects.List {
			if m.cfg.Projects.List[i].Path == mainPath {
				return &m.cfg.Projects.List[i]
			}
		}
	}

	return nil
}

// confirmThemeSelection saves the theme, reloads config, resets all theme
// switcher state, and returns a toast command. displayName is used in the toast.
func (m *Model) confirmThemeSelection(tc config.ThemeConfig, displayName string) tea.Cmd {
	scope := m.theme.Scope

	// Save before reset clears scope
	if err := m.saveTheme(tc, scope); err != nil {
		m.resetThemeSwitcher()
		m.updateContext()
		return func() tea.Msg {
			return ToastMsg{Message: "Theme applied (save failed)", Duration: 3 * time.Second, IsError: true}
		}
	}
	if cfg, err := config.Load(); err == nil {
		m.cfg = cfg
	}

	m.resetThemeSwitcher()
	m.updateContext()

	toastMsg := "Theme: " + displayName
	if scope == "project" {
		toastMsg += " (project)"
	} else {
		toastMsg += " (global)"
	}
	return func() tea.Msg {
		return ToastMsg{Message: toastMsg, Duration: 2 * time.Second}
	}
}

// saveTheme persists a ThemeConfig based on scope.
func (m *Model) saveTheme(tc config.ThemeConfig, scope string) error {
	if scope == "project" {
		projectPath := m.ui.WorkDir
		if pc := m.currentProjectConfig(); pc != nil {
			projectPath = pc.Path
		}
		return config.SaveProjectTheme(projectPath, &tc)
	}
	return config.SaveGlobalTheme(tc)
}

// copyProjectSetupPrompt copies an LLM-friendly prompt for configuring projects.
func (m *Model) copyProjectSetupPrompt() tea.Cmd {
	prompt := `Configure sidecar projects for me.

Add my code projects to ~/.config/sidecar/config.json using this format:

{
  "projects": {
    "list": [
      {"name": "short-name", "path": "~/code/project-path"}
    ]
  }
}

Rules:
- Use short, memorable names (1-2 words, lowercase, hyphens ok)
- Expand ~ to full home path if needed
- Only add directories that exist and contain code
- Merge with existing config if present

My code is located at: [TELL ME WHERE YOUR CODE DIRECTORIES ARE]`

	if err := clipboard.WriteAll(prompt); err != nil {
		return func() tea.Msg {
			return ToastMsg{Message: "Copy failed: " + err.Error(), Duration: 2 * time.Second}
		}
	}
	return func() tea.Msg {
		return ToastMsg{Message: "Copied LLM setup prompt", Duration: 2 * time.Second}
	}
}

// initProjectAdd initializes the project add sub-mode.
func (m *Model) initProjectAdd() {
	m.project.AddMode = true
	m.clearProjectAddModal()

	if m.project.Add == nil {
		m.project.Add = &projectAddState{}
	}
	m.project.Add.errorMessage = ""
	m.project.Add.themeSelected = ""

	nameInput := textinput.New()
	nameInput.Placeholder = "project-name"
	nameInput.CharLimit = 40
	nameInput.Width = 36
	nameInput.Focus()
	m.project.Add.nameInput = nameInput

	pathInput := textinput.New()
	pathInput.Placeholder = "~/code/project-path"
	pathInput.CharLimit = 200
	pathInput.Width = 36
	m.project.Add.pathInput = pathInput
}

// resetProjectAdd resets the project add sub-mode state.
func (m *Model) resetProjectAdd() {
	m.project.AddMode = false
	if m.project.Add != nil {
		m.project.Add.errorMessage = ""
		m.project.Add.themeSelected = ""
	}
	m.clearProjectAddModal()
	m.resetProjectAddThemePicker()
}

// initProjectAddThemePicker opens the theme picker sub-modal.
func (m *Model) initProjectAddThemePicker() {
	m.project.AddThemeMode = true
	ti := textinput.New()
	ti.Placeholder = "Filter themes..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 36
	m.project.AddThemeInput = ti
	m.project.AddThemeFiltered = append([]string{"(use global)"}, styles.ListThemes()...)
	m.project.AddThemeCursor = 0
	m.project.AddThemeScroll = 0
	m.project.AddCommunityMode = false
}

// resetProjectAddThemePicker closes the theme picker sub-modal.
func (m *Model) resetProjectAddThemePicker() {
	m.project.AddThemeMode = false
	m.project.AddCommunityMode = false
	m.project.AddThemeCursor = 0
	m.project.AddThemeScroll = 0
	m.project.AddCommunityCursor = 0
	m.project.AddCommunityScroll = 0
}

// previewProjectAddTheme previews the currently-selected built-in theme.
func (m *Model) previewProjectAddTheme() {
	if m.project.AddThemeCursor >= 0 && m.project.AddThemeCursor < len(m.project.AddThemeFiltered) {
		name := m.project.AddThemeFiltered[m.project.AddThemeCursor]
		if name == "(use global)" {
			resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
			theme.ApplyResolved(resolved)
		} else {
			theme.ApplyResolved(theme.ResolvedTheme{BaseName: name})
		}
	}
}

// previewProjectAddCommunity previews the currently-selected community theme.
func (m *Model) previewProjectAddCommunity() {
	if m.project.AddCommunityCursor >= 0 && m.project.AddCommunityCursor < len(m.project.AddCommunityList) {
		name := m.project.AddCommunityList[m.project.AddCommunityCursor]
		theme.ApplyResolved(theme.ResolvedTheme{BaseName: "default", CommunityName: name})
	}
}

// validateProjectAdd validates the project add form inputs.
// Returns an error message or empty string if valid.
func (m *Model) validateProjectAdd() string {
	if m.project.Add == nil {
		return "Name is required"
	}

	name := strings.TrimSpace(m.project.Add.nameInput.Value())
	path := strings.TrimSpace(m.project.Add.pathInput.Value())

	if name == "" {
		return "Name is required"
	}
	if path == "" {
		return "Path is required"
	}

	// Expand path for validation
	expanded := config.ExpandPath(path)

	// Check path exists and is a directory
	info, err := os.Stat(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return "Path does not exist"
		}
		return "Cannot access path"
	}
	if !info.IsDir() {
		return "Path is not a directory"
	}

	// Check for duplicate name or path
	for _, proj := range m.cfg.Projects.List {
		if strings.EqualFold(proj.Name, name) {
			return "Project name already exists"
		}
		if proj.Path == expanded {
			return "Project path already configured"
		}
	}

	return ""
}

// saveProjectAdd saves the new project to config and refreshes the list.
func (m *Model) saveProjectAdd() tea.Cmd {
	if m.project.Add == nil {
		return func() tea.Msg {
			return ToastMsg{Message: "Project add state missing", Duration: 3 * time.Second, IsError: true}
		}
	}

	name := strings.TrimSpace(m.project.Add.nameInput.Value())
	path := strings.TrimSpace(m.project.Add.pathInput.Value())

	// Build project config
	proj := config.ProjectConfig{
		Name: name,
		Path: config.ExpandPath(path),
	}

	// Add theme if user selected one
	if m.project.Add.themeSelected != "" && m.project.Add.themeSelected != "(use global)" {
		if community.GetScheme(m.project.Add.themeSelected) != nil {
			proj.Theme = &config.ThemeConfig{
				Name:      "default",
				Community: m.project.Add.themeSelected,
			}
		} else {
			proj.Theme = &config.ThemeConfig{
				Name: m.project.Add.themeSelected,
			}
		}
	}

	// Reload config from disk to avoid overwriting external changes
	cfg, err := config.Load()
	if err != nil {
		return func() tea.Msg {
			return ToastMsg{Message: "Failed to load config: " + err.Error(), Duration: 3 * time.Second, IsError: true}
		}
	}

	// Add project to fresh config
	cfg.Projects.List = append(cfg.Projects.List, proj)

	// Save to disk
	if err := config.Save(cfg); err != nil {
		return func() tea.Msg {
			return ToastMsg{Message: "Added project (save failed: " + err.Error() + ")", Duration: 3 * time.Second, IsError: true}
		}
	}

	// Update in-memory config
	m.cfg.Projects.List = cfg.Projects.List

	// Refresh the filtered list
	m.project.Filtered = m.cfg.Projects.List

	return func() tea.Msg {
		return ToastMsg{Message: fmt.Sprintf("Added project: %s", name), Duration: 3 * time.Second}
	}
}

// resetThemeSwitcher resets the theme switcher modal state.
func (m *Model) resetThemeSwitcher() {
	m.theme.Show = false
	m.theme.SelectedIdx = 0
	m.theme.Filtered = nil
	m.theme.Scope = ""
	m.theme.Original = themeEntry{}
	m.clearThemeSwitcherModal()
}

// clearThemeSwitcherModal clears the theme switcher modal state.
func (m *Model) clearThemeSwitcherModal() {
	m.theme.Modal = nil
	m.theme.ModalWidth = 0
	m.theme.MouseHandler = nil
}

// initIssueInput initializes the issue input modal.
func (m *Model) initIssueInput() {
	ti := textinput.New()
	ti.Placeholder = "Issue ID or search text"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50
	m.issue.InputModel = ti
	m.issue.InputModal = nil
	m.issue.InputModalWidth = 0
	m.issue.InputMouseHandler = mouse.NewHandler()
	m.issue.SearchResults = nil
	m.issue.SearchQuery = ""
	m.issue.SearchLoading = false
	m.issue.SearchCursor = -1
	m.issue.SearchScrollOffset = 0
	m.issue.SearchIncludeClosed = false
}

// resetIssueInput resets the issue input modal state.
func (m *Model) resetIssueInput() {
	m.issue.ShowInput = false
	m.issue.InputModal = nil
	m.issue.InputModalWidth = 0
	m.issue.InputMouseHandler = nil
	m.issue.SearchResults = nil
	m.issue.SearchQuery = ""
	m.issue.SearchLoading = false
	m.issue.SearchCursor = -1
	m.issue.SearchScrollOffset = 0
	m.issue.SearchIncludeClosed = false
}

// resetIssuePreview resets the issue preview modal state.
func (m *Model) resetIssuePreview() {
	m.issue.ShowPreview = false
	m.issue.PreviewData = nil
	m.issue.PreviewLoading = false
	m.issue.PreviewError = nil
	m.issue.PreviewModal = nil
	m.issue.PreviewModalWidth = 0
	m.issue.PreviewMouseHandler = nil
}

// backToIssueInput closes the preview and returns to the search modal
// with the previous query, results, and cursor intact.
func (m *Model) backToIssueInput() {
	m.resetIssuePreview()
	m.issue.ShowInput = true
	m.activeContext = "issue-input"
	m.issue.InputModel.Focus()
	m.issue.InputModal = nil
	m.issue.InputModalWidth = 0
	m.issue.InputMouseHandler = mouse.NewHandler()
}

// initThemeSwitcher initializes the theme switcher modal.
func (m *Model) initThemeSwitcher() {
	ti := textinput.New()
	ti.Placeholder = "Filter themes..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 54
	m.theme.Input = ti

	allEntries := buildUnifiedThemeList()
	m.theme.Filtered = allEntries
	m.theme.SelectedIdx = 0
	if m.currentProjectConfig() != nil {
		m.theme.Scope = "project"
	} else {
		m.theme.Scope = "global"
	}
	m.clearThemeSwitcherModal()

	// Determine original theme from config
	m.theme.Original = themeEntry{Name: "default", IsBuiltIn: true, ThemeKey: "default"}
	if freshCfg, err := config.Load(); err == nil {
		if freshCfg.UI.Theme.Community != "" {
			// Current theme is a community theme
			communityName := freshCfg.UI.Theme.Community
			m.theme.Original = themeEntry{Name: communityName, IsBuiltIn: false, ThemeKey: communityName}
		} else if freshCfg.UI.Theme.Name != "" {
			name := freshCfg.UI.Theme.Name
			displayName := name
			if t := styles.GetTheme(name); t.DisplayName != "" {
				displayName = t.DisplayName
			}
			m.theme.Original = themeEntry{Name: displayName, IsBuiltIn: true, ThemeKey: name}
		}
	}

	// Set cursor to current theme
	for i, entry := range m.theme.Filtered {
		if entry.IsBuiltIn == m.theme.Original.IsBuiltIn && entry.ThemeKey == m.theme.Original.ThemeKey {
			m.theme.SelectedIdx = i
			break
		}
	}
}

// previewThemeEntry applies the given theme entry for live preview.
func (m *Model) previewThemeEntry(entry themeEntry) {
	if entry.IsBuiltIn {
		m.applyThemeFromConfig(entry.ThemeKey)
	} else {
		theme.ApplyResolved(theme.ResolvedTheme{
			BaseName:      "default",
			CommunityName: entry.ThemeKey,
		})
	}
}

// applyThemeFromConfig applies a theme, using config overrides only if the
// saved config has that theme selected. This means live preview of other themes
// won't include user customizations (which is intentional - you want to see the
// base theme, not your customizations for a different theme).
func (m *Model) applyThemeFromConfig(themeName string) {
	freshCfg, err := config.Load()
	if err == nil && freshCfg.UI.Theme.Name == themeName {
		// Apply the saved theme with its full config (community + overrides)
		theme.ApplyResolved(theme.ResolvedTheme{
			BaseName:      themeName,
			CommunityName: freshCfg.UI.Theme.Community,
			Overrides:     freshCfg.UI.Theme.Overrides,
		})
	} else {
		styles.ApplyTheme(themeName)
	}
}
