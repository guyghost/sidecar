package styles

import (
	"sync/atomic"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles contains all color values (as lipgloss.Color) and all lipgloss.Style
// definitions. This is immutable once created - a new Styles instance is
// built for each theme change.
type Styles struct {
	// Color palette (as lipgloss.Color for direct use)
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// Text colors
	TextPrimary        lipgloss.Color
	TextSecondary      lipgloss.Color
	TextMuted          lipgloss.Color
	TextSubtle         lipgloss.Color
	TextSelectionColor lipgloss.Color

	// Background colors
	BgPrimary   lipgloss.Color
	BgSecondary lipgloss.Color
	BgTertiary  lipgloss.Color
	BgOverlay   lipgloss.Color

	// Border colors
	BorderNormal lipgloss.Color
	BorderActive lipgloss.Color
	BorderMuted  lipgloss.Color

	// Diff colors
	DiffAddFg    lipgloss.Color
	DiffRemoveFg lipgloss.Color
	DiffAddBg    lipgloss.Color
	DiffRemoveBg lipgloss.Color

	// Additional themeable colors
	TextHighlight         lipgloss.Color
	ButtonHoverColor      lipgloss.Color
	TabTextInactiveColor  lipgloss.Color
	LinkColor             lipgloss.Color
	ToastSuccessTextColor lipgloss.Color
	ToastErrorTextColor   lipgloss.Color

	// Danger button colors
	DangerLight  lipgloss.Color
	DangerDark   lipgloss.Color
	DangerBright lipgloss.Color
	DangerHover  lipgloss.Color
	TextInverse  lipgloss.Color

	// Scrollbar colors
	ScrollbarTrackColor lipgloss.Color
	ScrollbarThumbColor lipgloss.Color

	// Blame age gradient colors
	BlameAge1 lipgloss.Color
	BlameAge2 lipgloss.Color
	BlameAge3 lipgloss.Color
	BlameAge4 lipgloss.Color
	BlameAge5 lipgloss.Color

	// Panel styles
	PanelActive   lipgloss.Style
	PanelInactive lipgloss.Style
	PanelHeader   lipgloss.Style
	PanelNoBorder lipgloss.Style

	// Text styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	// WorktreeIndicator shows the current worktree branch in the header
	WorktreeIndicator lipgloss.Style

	Body    lipgloss.Style
	Muted   lipgloss.Style
	Subtle  lipgloss.Style
	Code    lipgloss.Style
	Link    lipgloss.Style
	KeyHint lipgloss.Style
	Logo    lipgloss.Style

	// Status indicator styles
	StatusStaged     lipgloss.Style
	StatusModified   lipgloss.Style
	ToastSuccess     lipgloss.Style
	ToastError       lipgloss.Style
	StatusUntracked  lipgloss.Style
	StatusDeleted    lipgloss.Style
	StatusInProgress lipgloss.Style
	StatusCompleted  lipgloss.Style
	StatusBlocked    lipgloss.Style
	StatusPending    lipgloss.Style

	// Note status indicator styles
	StatusArchived    lipgloss.Style
	StatusDeletedNote lipgloss.Style

	// List item styles
	ListItemNormal   lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemFocused  lipgloss.Style
	ListCursor       lipgloss.Style

	// Bar element styles (shared by header/footer)
	BarTitle      lipgloss.Style
	BarText       lipgloss.Style
	BarChip       lipgloss.Style
	BarChipActive lipgloss.Style

	// Tab styles
	TabTextActive   lipgloss.Style
	TabTextInactive lipgloss.Style

	// Diff line styles
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffContext lipgloss.Style
	DiffHeader  lipgloss.Style

	// File browser styles
	FileBrowserDir        lipgloss.Style
	FileBrowserFile       lipgloss.Style
	FileBrowserIgnored    lipgloss.Style
	FileBrowserLineNumber lipgloss.Style
	FileBrowserIcon       lipgloss.Style
	SearchMatch           lipgloss.Style
	SearchMatchCurrent    lipgloss.Style
	FuzzyMatchChar        lipgloss.Style
	QuickOpenItem         lipgloss.Style
	QuickOpenItemSelected lipgloss.Style
	PaletteEntry          lipgloss.Style
	PaletteEntrySelected  lipgloss.Style
	PaletteKey            lipgloss.Style
	TextSelection         lipgloss.Style

	// Footer and header
	Footer lipgloss.Style
	Header lipgloss.Style

	// Modal styles
	ModalOverlay lipgloss.Style
	ModalBox     lipgloss.Style
	ModalTitle   lipgloss.Style

	// Button styles
	Button              lipgloss.Style
	ButtonFocused       lipgloss.Style
	ButtonHover         lipgloss.Style
	ButtonDanger        lipgloss.Style
	ButtonDangerFocused lipgloss.Style
	ButtonDangerHover   lipgloss.Style
}

// NewStyles creates a complete Styles instance from a color palette.
// All lipgloss styles are built from palette's color values.
func NewStyles(p Palette) *Styles {
	return &Styles{
		// Color palette
		Primary:   lipgloss.Color(p.Primary),
		Secondary: lipgloss.Color(p.Secondary),
		Accent:    lipgloss.Color(p.Accent),

		Success: lipgloss.Color(p.Success),
		Warning: lipgloss.Color(p.Warning),
		Error:   lipgloss.Color(p.Error),
		Info:    lipgloss.Color(p.Info),

		TextPrimary:        lipgloss.Color(p.TextPrimary),
		TextSecondary:      lipgloss.Color(p.TextSecondary),
		TextMuted:          lipgloss.Color(p.TextMuted),
		TextSubtle:         lipgloss.Color(p.TextSubtle),
		TextSelectionColor: lipgloss.Color(p.TextSelectionColor),

		BgPrimary:   lipgloss.Color(p.BgPrimary),
		BgSecondary: lipgloss.Color(p.BgSecondary),
		BgTertiary:  lipgloss.Color(p.BgTertiary),
		BgOverlay:   lipgloss.Color(p.BgOverlay),

		BorderNormal: lipgloss.Color(p.BorderNormal),
		BorderActive: lipgloss.Color(p.BorderActive),
		BorderMuted:  lipgloss.Color(p.BorderMuted),

		DiffAddFg:    lipgloss.Color(p.DiffAddFg),
		DiffRemoveFg: lipgloss.Color(p.DiffRemoveFg),
		DiffAddBg:    lipgloss.Color(p.DiffAddBg),
		DiffRemoveBg: lipgloss.Color(p.DiffRemoveBg),

		TextHighlight:         lipgloss.Color(p.TextHighlight),
		ButtonHoverColor:      lipgloss.Color(p.ButtonHoverColor),
		TabTextInactiveColor:  lipgloss.Color(p.TabTextInactiveColor),
		LinkColor:             lipgloss.Color(p.LinkColor),
		ToastSuccessTextColor: lipgloss.Color(p.ToastSuccessTextColor),
		ToastErrorTextColor:   lipgloss.Color(p.ToastErrorTextColor),

		DangerLight:  lipgloss.Color(p.DangerLight),
		DangerDark:   lipgloss.Color(p.DangerDark),
		DangerBright: lipgloss.Color(p.DangerBright),
		DangerHover:  lipgloss.Color(p.DangerHover),
		TextInverse:  lipgloss.Color(p.TextInverse),

		ScrollbarTrackColor: lipgloss.Color(p.ScrollbarTrackColor),
		ScrollbarThumbColor: lipgloss.Color(p.ScrollbarThumbColor),

		BlameAge1: lipgloss.Color(p.BlameAge1),
		BlameAge2: lipgloss.Color(p.BlameAge2),
		BlameAge3: lipgloss.Color(p.BlameAge3),
		BlameAge4: lipgloss.Color(p.BlameAge4),
		BlameAge5: lipgloss.Color(p.BlameAge5),

		// Panel styles
		PanelActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(p.BorderActive)).
			Padding(0, 1),

		PanelInactive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(p.BorderNormal)).
			Padding(0, 1),

		PanelHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.TextPrimary)).
			MarginBottom(1),

		PanelNoBorder: lipgloss.NewStyle().
			Padding(0, 1),

		// Text styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.TextPrimary)),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextHighlight)),

		// WorktreeIndicator shows the current worktree branch in the header
		WorktreeIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Warning)).
			Bold(true),

		Body: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSubtle)),

		Code: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Accent)),

		Link: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.LinkColor)).
			Underline(true),

		KeyHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Background(lipgloss.Color(p.BgTertiary)).
			Padding(0, 1),

		Logo: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Primary)).
			Bold(true),

		// Status indicator styles
		StatusStaged: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Success)).
			Bold(true),

		StatusModified: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Warning)).
			Bold(true),

		// Toast styles for status messages
		ToastSuccess: lipgloss.NewStyle().
			Background(lipgloss.Color(p.Success)).
			Foreground(lipgloss.Color(p.ToastSuccessTextColor)).
			Bold(true).
			Padding(0, 1),

		ToastError: lipgloss.NewStyle().
			Background(lipgloss.Color(p.Error)).
			Foreground(lipgloss.Color(p.ToastErrorTextColor)).
			Bold(true).
			Padding(0, 1),

		StatusUntracked: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		StatusDeleted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Error)).
			Bold(true),

		StatusInProgress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Info)).
			Bold(true),

		StatusCompleted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Success)),

		StatusBlocked: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Error)),

		StatusPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		// Note status indicator styles
		StatusArchived: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Info)),

		StatusDeletedNote: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Error)),

		// List item styles
		ListItemNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)),

		ListItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSelectionColor)).
			Background(lipgloss.Color(p.BgTertiary)),

		ListItemFocused: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Background(lipgloss.Color(p.Primary)),

		ListCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Primary)).
			Bold(true),

		// Bar element styles (shared by header/footer)
		BarTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Bold(true),

		BarText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		BarChip: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Background(lipgloss.Color(p.BgTertiary)).
			Padding(0, 1),

		BarChipActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 1).
			Bold(true),

		// TabTextActive is the text color for active tabs
		TabTextActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Bold(true),

		// TabTextInactive is the text color for inactive tabs
		TabTextInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TabTextInactiveColor)),

		// Diff line styles
		DiffAdd: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Success)),

		DiffRemove: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Error)),

		DiffContext: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		DiffHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Info)).
			Bold(true),

		// File browser styles
		// Directory names - bold blue
		FileBrowserDir: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Secondary)).
			Bold(true),

		// Regular file names
		FileBrowserFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)),

		// Gitignored files - muted/dimmed
		FileBrowserIgnored: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSubtle)),

		// Line numbers in preview
		FileBrowserLineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Width(5).
			AlignHorizontal(lipgloss.Right),

		// Tree icons (>, +)
		FileBrowserIcon: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		// Content search match highlighting
		SearchMatch: lipgloss.NewStyle().
			Background(lipgloss.Color(p.Warning)),

		SearchMatchCurrent: lipgloss.NewStyle().
			Background(lipgloss.Color(p.Primary)).
			Foreground(lipgloss.Color(p.TextPrimary)),

		// Fuzzy match character highlighting (bold in result list)
		FuzzyMatchChar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Primary)).
			Bold(true),

		// Quick open result row (normal)
		QuickOpenItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)),

		// Quick open result row (selected)
		QuickOpenItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSelectionColor)).
			Background(lipgloss.Color(p.BgTertiary)),

		// Palette entry styles (reusable for modals)
		PaletteEntry: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)),

		PaletteEntrySelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSelectionColor)).
			Background(lipgloss.Color(p.BgTertiary)),

		PaletteKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Background(lipgloss.Color(p.BgTertiary)).
			Padding(0, 1),

		// Text selection for preview pane drag selection
		TextSelection: lipgloss.NewStyle().
			Background(lipgloss.Color(p.BgTertiary)).
			Foreground(lipgloss.Color(p.TextSelectionColor)),

		// Footer and header
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Background(lipgloss.Color(p.BgSecondary)),

		Header: lipgloss.NewStyle().
			Background(lipgloss.Color(p.BgSecondary)),

		// Modal styles
		ModalOverlay: lipgloss.NewStyle().
			Background(lipgloss.Color(p.BgOverlay)),

		ModalBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(p.Primary)).
			Background(lipgloss.Color(p.BgSecondary)).
			Padding(1, 2),

		ModalTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Bold(true).
			MarginBottom(1),

		// Button styles
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextSecondary)).
			Background(lipgloss.Color(p.BgTertiary)).
			Padding(0, 2),

		ButtonFocused: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 2).
			Bold(true),

		ButtonHover: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextPrimary)).
			Background(lipgloss.Color(p.ButtonHoverColor)).
			Padding(0, 2),

		// Danger button styles (for destructive actions like delete)
		ButtonDanger: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.DangerLight)).
			Background(lipgloss.Color(p.DangerDark)).
			Padding(0, 2),

		ButtonDangerFocused: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextInverse)).
			Background(lipgloss.Color(p.DangerBright)).
			Padding(0, 2).
			Bold(true),

		ButtonDangerHover: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextInverse)).
			Background(lipgloss.Color(p.DangerHover)).
			Padding(0, 2),
	}
}

// Current holds the active Styles instance. This is atomically swappable
// to enable thread-safe theme switching without stopping the TUI.
var Current atomic.Pointer[Styles]

// Initialize sets up the initial styles. This must be called once at startup.
// It is safe to call multiple times (idempotent).
func Initialize() {
	if Current.Load() == nil {
		Current.Store(NewStyles(DefaultPalette()))
	}
}

// GetStyles returns the current active Styles instance.
// This is thread-safe and can be called from anywhere.
func GetStyles() *Styles {
	if s := Current.Load(); s != nil {
		return s
	}
	// Fallback: initialize if not yet done
	Initialize()
	return Current.Load()
}

// Backward-compatible package-level variables.
// These are updated atomically by ApplyThemeColors and can be read directly
// by existing code that references styles.Primary, styles.PanelActive, etc.

// Color palette - default dark theme
var (
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// Text colors
	TextPrimary        lipgloss.Color
	TextSecondary      lipgloss.Color
	TextMuted          lipgloss.Color
	TextSubtle         lipgloss.Color
	TextSelectionColor lipgloss.Color // Text on selection backgrounds (BgTertiary)

	// Background colors
	BgPrimary   lipgloss.Color
	BgSecondary lipgloss.Color
	BgTertiary  lipgloss.Color
	BgOverlay   lipgloss.Color

	// Border colors
	BorderNormal lipgloss.Color
	BorderActive lipgloss.Color
	BorderMuted  lipgloss.Color

	// Diff foreground colors (also updated by ApplyTheme)
	DiffAddFg    lipgloss.Color
	DiffRemoveFg lipgloss.Color

	// Additional themeable colors
	TextHighlight         lipgloss.Color // For subtitle, special text
	ButtonHoverColor      lipgloss.Color // Button hover background
	TabTextInactiveColor  lipgloss.Color // Inactive tab text
	LinkColor             lipgloss.Color // Hyperlink color
	ToastSuccessTextColor lipgloss.Color // Toast success foreground
	ToastErrorTextColor   lipgloss.Color // Toast error foreground

	// Danger button colors
	DangerLight  lipgloss.Color // Light red text
	DangerDark   lipgloss.Color // Dark red background
	DangerBright lipgloss.Color // Bright red focused bg
	DangerHover  lipgloss.Color // Darker red hover bg
	TextInverse  lipgloss.Color // Inverse/contrast text

	// Scrollbar colors (default to TextSubtle/TextMuted)
	ScrollbarTrackColor lipgloss.Color // Same as TextSubtle
	ScrollbarThumbColor lipgloss.Color // Same as TextMuted

	// Blame age gradient colors
	BlameAge1 lipgloss.Color // < 1 week
	BlameAge2 lipgloss.Color // < 1 month
	BlameAge3 lipgloss.Color // < 3 months
	BlameAge4 lipgloss.Color // < 6 months
	BlameAge5 lipgloss.Color // < 1 year

	// Third-party theme names (updated by ApplyTheme)
	CurrentSyntaxTheme   = "monokai"
	CurrentMarkdownTheme = "dark"
)

// Tab theme state (updated by ApplyTheme)
var (
	CurrentTabStyle  = "rainbow"
	CurrentTabColors = []RGB{{220, 60, 60}, {60, 220, 60}, {60, 60, 220}, {156, 60, 220}} // Default rainbow
)

// PillTabsEnabled enables rounded pill-style tabs (requires Nerd Font)
var PillTabsEnabled = false

// Panel styles
var (
	// Active panel with highlighted border
	PanelActive lipgloss.Style

	// Inactive panel with subtle border
	PanelInactive lipgloss.Style

	// Panel header
	PanelHeader lipgloss.Style

	// Panel with no border
	PanelNoBorder lipgloss.Style
)

// Text styles
var (
	Title lipgloss.Style

	Subtitle lipgloss.Style

	// WorktreeIndicator shows the current worktree branch in the header
	WorktreeIndicator lipgloss.Style

	Body lipgloss.Style

	Muted lipgloss.Style

	Subtle lipgloss.Style

	Code lipgloss.Style

	Link lipgloss.Style

	KeyHint lipgloss.Style

	Logo lipgloss.Style
)

// Status indicator styles
var (
	StatusStaged lipgloss.Style

	StatusModified lipgloss.Style

	// Toast styles for status messages
	ToastSuccess lipgloss.Style

	ToastError lipgloss.Style

	StatusUntracked lipgloss.Style

	StatusDeleted lipgloss.Style

	StatusInProgress lipgloss.Style

	StatusCompleted lipgloss.Style

	StatusBlocked lipgloss.Style

	StatusPending lipgloss.Style

	// Note status indicator styles
	StatusArchived lipgloss.Style

	StatusDeletedNote lipgloss.Style
)

// List item styles
var (
	ListItemNormal lipgloss.Style

	ListItemSelected lipgloss.Style

	ListItemFocused lipgloss.Style

	ListCursor lipgloss.Style
)

// Bar element styles (shared by header/footer)
var (
	BarTitle lipgloss.Style

	BarText lipgloss.Style

	BarChip lipgloss.Style

	BarChipActive lipgloss.Style
)

// TabTextActive is the text color for active tabs
var TabTextActive lipgloss.Style

// TabTextInactive is the text color for inactive tabs
var TabTextInactive lipgloss.Style

// Diff line styles
var (
	DiffAdd lipgloss.Style

	DiffRemove lipgloss.Style

	DiffContext lipgloss.Style

	DiffHeader lipgloss.Style

	// Subtle diff backgrounds for syntax-highlighted lines
	DiffAddBg    lipgloss.Color
	DiffRemoveBg lipgloss.Color
)

// File browser styles
var (
	// Directory names - bold blue
	FileBrowserDir lipgloss.Style

	// Regular file names
	FileBrowserFile lipgloss.Style

	// Gitignored files - muted/dimmed
	FileBrowserIgnored lipgloss.Style

	// Line numbers in preview
	FileBrowserLineNumber lipgloss.Style

	// Tree icons (>, +)
	FileBrowserIcon lipgloss.Style

	// Content search match highlighting
	SearchMatch lipgloss.Style

	SearchMatchCurrent lipgloss.Style

	// Fuzzy match character highlighting (bold in result list)
	FuzzyMatchChar lipgloss.Style

	// Quick open result row (normal)
	QuickOpenItem lipgloss.Style

	// Quick open result row (selected)
	QuickOpenItemSelected lipgloss.Style

	// Palette entry styles (reusable for modals)
	PaletteEntry lipgloss.Style

	PaletteEntrySelected lipgloss.Style

	PaletteKey lipgloss.Style

	// Text selection for preview pane drag selection
	TextSelection lipgloss.Style
)

// Footer and header
var (
	Footer lipgloss.Style

	Header lipgloss.Style
)

// Modal styles
var (
	ModalOverlay lipgloss.Style

	ModalBox lipgloss.Style

	ModalTitle lipgloss.Style
)

// Button styles
var (
	Button lipgloss.Style

	ButtonFocused lipgloss.Style

	ButtonHover lipgloss.Style

	// Danger button styles (for destructive actions like delete)
	ButtonDanger lipgloss.Style

	ButtonDangerFocused lipgloss.Style

	ButtonDangerHover lipgloss.Style
)

// syncFromStyles copies all colors and styles from a Styles instance
// to the package-level variables. This is called by ApplyThemeColors
// to maintain backward compatibility with existing code.
func syncFromStyles(s *Styles) {
	// Colors
	Primary = s.Primary
	Secondary = s.Secondary
	Accent = s.Accent
	Success = s.Success
	Warning = s.Warning
	Error = s.Error
	Info = s.Info
	TextPrimary = s.TextPrimary
	TextSecondary = s.TextSecondary
	TextMuted = s.TextMuted
	TextSubtle = s.TextSubtle
	TextSelectionColor = s.TextSelectionColor
	BgPrimary = s.BgPrimary
	BgSecondary = s.BgSecondary
	BgTertiary = s.BgTertiary
	BgOverlay = s.BgOverlay
	BorderNormal = s.BorderNormal
	BorderActive = s.BorderActive
	BorderMuted = s.BorderMuted
	DiffAddFg = s.DiffAddFg
	DiffRemoveFg = s.DiffRemoveFg
	DiffAddBg = s.DiffAddBg
	DiffRemoveBg = s.DiffRemoveBg
	TextHighlight = s.TextHighlight
	ButtonHoverColor = s.ButtonHoverColor
	TabTextInactiveColor = s.TabTextInactiveColor
	LinkColor = s.LinkColor
	ToastSuccessTextColor = s.ToastSuccessTextColor
	ToastErrorTextColor = s.ToastErrorTextColor
	DangerLight = s.DangerLight
	DangerDark = s.DangerDark
	DangerBright = s.DangerBright
	DangerHover = s.DangerHover
	TextInverse = s.TextInverse
	ScrollbarTrackColor = s.ScrollbarTrackColor
	ScrollbarThumbColor = s.ScrollbarThumbColor
	BlameAge1 = s.BlameAge1
	BlameAge2 = s.BlameAge2
	BlameAge3 = s.BlameAge3
	BlameAge4 = s.BlameAge4
	BlameAge5 = s.BlameAge5

	// Styles
	PanelActive = s.PanelActive
	PanelInactive = s.PanelInactive
	PanelHeader = s.PanelHeader
	PanelNoBorder = s.PanelNoBorder
	Title = s.Title
	Subtitle = s.Subtitle
	WorktreeIndicator = s.WorktreeIndicator
	Body = s.Body
	Muted = s.Muted
	Subtle = s.Subtle
	Code = s.Code
	Link = s.Link
	KeyHint = s.KeyHint
	Logo = s.Logo
	StatusStaged = s.StatusStaged
	StatusModified = s.StatusModified
	ToastSuccess = s.ToastSuccess
	ToastError = s.ToastError
	StatusUntracked = s.StatusUntracked
	StatusDeleted = s.StatusDeleted
	StatusInProgress = s.StatusInProgress
	StatusCompleted = s.StatusCompleted
	StatusBlocked = s.StatusBlocked
	StatusPending = s.StatusPending
	StatusArchived = s.StatusArchived
	StatusDeletedNote = s.StatusDeletedNote
	ListItemNormal = s.ListItemNormal
	ListItemSelected = s.ListItemSelected
	ListItemFocused = s.ListItemFocused
	ListCursor = s.ListCursor
	BarTitle = s.BarTitle
	BarText = s.BarText
	BarChip = s.BarChip
	BarChipActive = s.BarChipActive
	TabTextActive = s.TabTextActive
	TabTextInactive = s.TabTextInactive
	DiffAdd = s.DiffAdd
	DiffRemove = s.DiffRemove
	DiffContext = s.DiffContext
	DiffHeader = s.DiffHeader
	FileBrowserDir = s.FileBrowserDir
	FileBrowserFile = s.FileBrowserFile
	FileBrowserIgnored = s.FileBrowserIgnored
	FileBrowserLineNumber = s.FileBrowserLineNumber
	FileBrowserIcon = s.FileBrowserIcon
	SearchMatch = s.SearchMatch
	SearchMatchCurrent = s.SearchMatchCurrent
	FuzzyMatchChar = s.FuzzyMatchChar
	QuickOpenItem = s.QuickOpenItem
	QuickOpenItemSelected = s.QuickOpenItemSelected
	PaletteEntry = s.PaletteEntry
	PaletteEntrySelected = s.PaletteEntrySelected
	PaletteKey = s.PaletteKey
	TextSelection = s.TextSelection
	Footer = s.Footer
	Header = s.Header
	ModalOverlay = s.ModalOverlay
	ModalBox = s.ModalBox
	ModalTitle = s.ModalTitle
	Button = s.Button
	ButtonFocused = s.ButtonFocused
	ButtonHover = s.ButtonHover
	ButtonDanger = s.ButtonDanger
	ButtonDangerFocused = s.ButtonDangerFocused
	ButtonDangerHover = s.ButtonDangerHover
}

// parseTabColors converts hex color strings to RGB values for tab rendering
func parseTabColors(hexColors []string) []RGB {
	if len(hexColors) == 0 {
		// Return default rainbow colors
		return []RGB{{220, 60, 60}, {60, 220, 60}, {60, 60, 220}, {156, 60, 220}}
	}

	colors := make([]RGB, len(hexColors))
	for i, hex := range hexColors {
		colors[i] = HexToRGB(hex)
	}
	return colors
}

// ColorVars tracks initialization time to detect stale cached values
var initTime time.Time

func init() {
	Initialize()
	syncFromStyles(GetStyles())
	initTime = time.Now()
}

// IsStale returns true if the styles may be stale (for debugging/testing).
func IsStale() bool {
	return time.Since(initTime) > time.Hour
}

// RenderTab renders a tab label using the current tab theme.
// tabIndex is the 0-based index of this tab, totalTabs is the total count.
// If isPreview is true, the label is rendered in italic to indicate an ephemeral preview tab.
func RenderTab(label string, tabIndex, totalTabs int, isActive bool, isPreview bool) string {
	style := CurrentTabStyle
	colors := CurrentTabColors

	// Check if style is a preset name
	if preset := GetTabPreset(style); preset != nil {
		style = preset.Style
		if len(preset.Colors) > 0 {
			colors = parseTabColors(preset.Colors)
		}
	}

	switch style {
	case "gradient":
		return renderGradientTab(label, tabIndex, totalTabs, isActive, isPreview, colors)
	case "per-tab":
		return renderPerTabColor(label, tabIndex, isActive, isPreview, colors)
	case "solid":
		return renderSolidTab(label, isActive, isPreview)
	case "minimal":
		return renderMinimalTab(label, isActive, isPreview)
	default:
		// Default to gradient
		return renderGradientTab(label, tabIndex, totalTabs, isActive, isPreview, colors)
	}
}

// RenderGradientTab renders a tab label with a gradient background.
// Kept for backwards compatibility - delegates to RenderTab.
func RenderGradientTab(label string, tabIndex, totalTabs int, isActive bool) string {
	return renderGradientTab(label, tabIndex, totalTabs, isActive, false, CurrentTabColors)
}

func tabTextColor(isActive bool, backgrounds []RGB) lipgloss.Color {
	minTarget := 3.5
	candidates := []lipgloss.Color{TextSecondary, TextPrimary, TextMuted}
	if isActive {
		minTarget = 4.5
		candidates = []lipgloss.Color{TextPrimary, TextSecondary, TextMuted}
	}

	best := candidates[0]
	bestRatio := minContrastRatio(colorToRGB(best), backgrounds)
	for _, candidate := range candidates[1:] {
		if ratio := minContrastRatio(colorToRGB(candidate), backgrounds); ratio > bestRatio {
			best = candidate
			bestRatio = ratio
		}
	}

	if bestRatio < minTarget {
		for _, candidate := range []lipgloss.Color{lipgloss.Color("#000000"), lipgloss.Color("#ffffff")} {
			if ratio := minContrastRatio(colorToRGB(candidate), backgrounds); ratio > bestRatio {
				best = candidate
				bestRatio = ratio
			}
		}
	}

	return best
}

func colorToRGB(c lipgloss.Color) RGB {
	return HexToRGB(string(c))
}

// Pill tab cap characters (Powerline/Nerd Font)
const (
	pillLeftCap  = "\ue0b6" //
	pillRightCap = "\ue0b4" //
)

// RenderPill renders text with pill-shaped caps when NerdFontsEnabled (PillTabsEnabled) is true.
// fg is the text color, bg is the pill background, outerBg is the surrounding background.
// If outerBg is empty, defaults to BgSecondary.
func RenderPill(text string, fg, bg, outerBg lipgloss.Color) string {
	if outerBg == "" {
		outerBg = BgSecondary
	}

	style := lipgloss.NewStyle().Foreground(fg).Background(bg).Padding(0, 1)
	content := style.Render(text)

	if !PillTabsEnabled {
		return content
	}

	leftCap := lipgloss.NewStyle().Foreground(bg).Background(outerBg).Render(pillLeftCap)
	rightCap := lipgloss.NewStyle().Foreground(bg).Background(outerBg).Render(pillRightCap)

	// Re-render content without horizontal padding since caps provide visual space
	style = lipgloss.NewStyle().Foreground(fg).Background(bg)
	content = style.Render(" " + text + " ")

	return leftCap + content + rightCap
}

// RenderPillWithStyle renders text with pill-shaped caps using the provided lipgloss.Style.
// The style's background color is used for the pill caps.
// outerBg is the surrounding background; if empty, defaults to BgSecondary.
func RenderPillWithStyle(text string, style lipgloss.Style, outerBg lipgloss.Color) string {
	if outerBg == "" {
		outerBg = BgSecondary
	}

	// Get background from style for pill caps
	bg, _ := style.GetBackground().(lipgloss.Color)
	if bg == "" {
		bg = BgTertiary // fallback
	}

	content := style.Render(text)

	if !PillTabsEnabled {
		return content
	}

	leftCap := lipgloss.NewStyle().Foreground(bg).Background(outerBg).Render(pillLeftCap)
	rightCap := lipgloss.NewStyle().Foreground(bg).Background(outerBg).Render(pillRightCap)

	// Re-render content - adjust padding for pill caps
	// Remove existing padding and add single space padding
	innerStyle := style.Padding(0, 0)
	content = innerStyle.Render(" " + text + " ")

	return leftCap + content + rightCap
}

// renderGradientTab renders a tab with per-character gradient coloring.
func renderGradientTab(label string, tabIndex, totalTabs int, isActive bool, isPreview bool, colors []RGB) string {
	if totalTabs == 0 {
		totalTabs = 1
	}
	if tabIndex < 0 {
		tabIndex = 0
	}

	// Calculate position in the gradient (0.0 to 1.0 across all tabs)
	tabWidth := 1.0 / float64(totalTabs)
	startPos := float64(tabIndex) * tabWidth
	endPos := startPos + tabWidth

	// Add padding to label (less if using pill caps since they add visual width)
	padded := " " + label + " "
	if !PillTabsEnabled {
		padded = "  " + label + "  "
	}
	chars := []rune(padded)
	backgrounds := make([]RGB, len(chars))
	result := ""

	for i := range chars {
		// Position within the gradient for this character
		charPos := startPos + (endPos-startPos)*float64(i)/float64(len(chars))

		// Get interpolated color
		r, g, b := interpolateColors(charPos, colors)

		// Mute colors for inactive tabs
		if !isActive {
			r = uint8(float64(r)*0.35 + 30)
			g = uint8(float64(g)*0.35 + 30)
			b = uint8(float64(b)*0.35 + 30)
		}
		backgrounds[i] = RGB{float64(r), float64(g), float64(b)}
	}

	// Left pill cap: foreground = first char's bg, background = header bg
	if PillTabsEnabled {
		leftBg := lipgloss.Color(RGBToHex(backgrounds[0]))
		leftCap := lipgloss.NewStyle().Foreground(leftBg).Background(BgSecondary).Render(pillLeftCap)
		result += leftCap
	}

	textColor := tabTextColor(isActive, backgrounds)
	for i, ch := range chars {
		bg := lipgloss.Color(RGBToHex(backgrounds[i]))
		var style lipgloss.Style
		if isActive {
			style = lipgloss.NewStyle().Background(bg).Foreground(textColor).Bold(true)
		} else {
			style = lipgloss.NewStyle().Background(bg).Foreground(textColor)
		}
		if isPreview {
			style = style.Italic(true)
		}
		result += style.Render(string(ch))
	}

	// Right pill cap: foreground = last char's bg, background = header bg
	if PillTabsEnabled {
		rightBg := lipgloss.Color(RGBToHex(backgrounds[len(backgrounds)-1]))
		rightCap := lipgloss.NewStyle().Foreground(rightBg).Background(BgSecondary).Render(pillRightCap)
		result += rightCap
	}

	return result
}

// renderPerTabColor renders a tab with a single solid color from the colors array.
func renderPerTabColor(label string, tabIndex int, isActive bool, isPreview bool, colors []RGB) string {
	if len(colors) == 0 {
		return renderSolidTab(label, isActive, isPreview)
	}

	// Guard against negative tabIndex (Go modulo returns negative for negative input)
	if tabIndex < 0 {
		tabIndex = 0
	}

	// Get color for this tab (cycle through available colors)
	color := colors[tabIndex%len(colors)]
	r, g, b := uint8(color.R), uint8(color.G), uint8(color.B)

	// Mute colors for inactive tabs
	if !isActive {
		r = uint8(float64(r)*0.35 + 30)
		g = uint8(float64(g)*0.35 + 30)
		b = uint8(float64(b)*0.35 + 30)
	}

	bgColor := RGB{float64(r), float64(g), float64(b)}
	textColor := tabTextColor(isActive, []RGB{bgColor})
	bg := lipgloss.Color(RGBToHex(bgColor))

	padded := " " + label + " "
	if !PillTabsEnabled {
		padded = "  " + label + "  "
	}

	var style lipgloss.Style
	if isActive {
		style = lipgloss.NewStyle().Background(bg).Foreground(textColor).Bold(true)
	} else {
		style = lipgloss.NewStyle().Background(bg).Foreground(textColor)
	}
	if isPreview {
		style = style.Italic(true)
	}

	if PillTabsEnabled {
		leftCap := lipgloss.NewStyle().Foreground(bg).Background(BgSecondary).Render(pillLeftCap)
		rightCap := lipgloss.NewStyle().Foreground(bg).Background(BgSecondary).Render(pillRightCap)
		return leftCap + style.Render(padded) + rightCap
	}
	return style.Render(padded)
}

// renderSolidTab renders a tab with the theme's primary/tertiary colors.
func renderSolidTab(label string, isActive bool, isPreview bool) string {
	padded := " " + label + " "
	if !PillTabsEnabled {
		padded = "  " + label + "  "
	}

	var bg lipgloss.Color
	if isActive {
		bg = Primary
	} else {
		bg = BgTertiary
	}

	textColor := tabTextColor(isActive, []RGB{colorToRGB(bg)})
	style := lipgloss.NewStyle().Background(bg).Foreground(textColor)
	if isActive {
		style = style.Bold(true)
	}
	if isPreview {
		style = style.Italic(true)
	}

	if PillTabsEnabled {
		leftCap := lipgloss.NewStyle().Foreground(bg).Background(BgSecondary).Render(pillLeftCap)
		rightCap := lipgloss.NewStyle().Foreground(bg).Background(BgSecondary).Render(pillRightCap)
		return leftCap + style.Render(padded) + rightCap
	}
	return style.Render(padded)
}

// renderMinimalTab renders a tab with no background, using underline for active.
func renderMinimalTab(label string, isActive bool, isPreview bool) string {
	padded := "  " + label + "  "

	var style lipgloss.Style
	if isActive {
		style = lipgloss.NewStyle().Foreground(Primary).Bold(true).Underline(true)
	} else {
		style = lipgloss.NewStyle().Foreground(TextMuted)
	}
	if isPreview {
		style = style.Italic(true)
	}

	return style.Render(padded)
}

// interpolateColors returns RGB for a position 0.0-1.0 across the color array
func interpolateColors(pos float64, colors []RGB) (uint8, uint8, uint8) {
	if len(colors) < 2 {
		if len(colors) == 1 {
			return uint8(colors[0].R), uint8(colors[0].G), uint8(colors[0].B)
		}
		return 128, 128, 128
	}

	// Scale position to color index
	scaled := pos * float64(len(colors)-1)
	idx := int(scaled)
	if idx >= len(colors)-1 {
		idx = len(colors) - 2
	}
	frac := scaled - float64(idx)

	// Interpolate between adjacent colors
	c1, c2 := colors[idx], colors[idx+1]
	r := uint8(c1.R + frac*(c2.R-c1.R))
	g := uint8(c1.G + frac*(c2.G-c1.G))
	b := uint8(c1.B + frac*(c2.B-c1.B))

	return r, g, b
}
