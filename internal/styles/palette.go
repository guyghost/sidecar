package styles

// Palette contains all color definitions for a theme.
// Colors are stored as hex strings and converted to lipgloss.Color when needed.
type Palette struct {
	// Primary colors
	Primary   string
	Secondary string
	Accent    string

	// Status colors
	Success string
	Warning string
	Error   string
	Info    string

	// Text colors
	TextPrimary        string
	TextSecondary      string
	TextMuted          string
	TextSubtle         string
	TextSelectionColor string // Text on selection backgrounds

	// Background colors
	BgPrimary   string
	BgSecondary string
	BgTertiary  string
	BgOverlay   string

	// Border colors
	BorderNormal string
	BorderActive string
	BorderMuted  string

	// Diff colors
	DiffAddFg    string
	DiffRemoveFg string
	DiffAddBg    string
	DiffRemoveBg string

	// Additional themeable colors
	TextHighlight         string
	ButtonHoverColor      string
	TabTextInactiveColor  string
	LinkColor             string
	ToastSuccessTextColor string
	ToastErrorTextColor   string

	// Danger button colors
	DangerLight  string
	DangerDark   string
	DangerBright string
	DangerHover  string
	TextInverse  string

	// Scrollbar colors
	ScrollbarTrackColor string
	ScrollbarThumbColor string

	// Blame age gradient colors
	BlameAge1 string
	BlameAge2 string
	BlameAge3 string
	BlameAge4 string
	BlameAge5 string

	// Third-party theme names
	SyntaxTheme   string
	MarkdownTheme string
}

// DefaultPalette returns the default dark theme color palette.
func DefaultPalette() Palette {
	return Palette{
		// Primary colors
		Primary:   "#7C3AED", // Purple
		Secondary: "#3B82F6", // Blue
		Accent:    "#F59E0B", // Amber

		// Status colors
		Success: "#10B981", // Green
		Warning: "#F59E0B", // Amber
		Error:   "#EF4444", // Red
		Info:    "#3B82F6", // Blue

		// Text colors
		TextPrimary:        "#F9FAFB",
		TextSecondary:      "#9CA3AF",
		TextMuted:          "#6B7280",
		TextSubtle:         "#4B5563",
		TextSelectionColor: "#F9FAFB", // Text on selection backgrounds

		// Background colors
		BgPrimary:   "#111827",
		BgSecondary: "#1F2937",
		BgTertiary:  "#374151",
		BgOverlay:   "#00000080",

		// Border colors
		BorderNormal: "#374151",
		BorderActive: "#7C3AED",
		BorderMuted:  "#1F2937",

		// Diff colors
		DiffAddFg:    "#10B981",
		DiffRemoveFg: "#EF4444",
		DiffAddBg:    "#0D2818", // Very subtle dark green
		DiffRemoveBg: "#2D1A1A", // Very subtle dark red

		// Additional themeable colors
		TextHighlight:         "#E5E7EB", // For subtitle, special text
		ButtonHoverColor:      "#9D174D", // Button hover background
		TabTextInactiveColor:  "#1a1a1a", // Inactive tab text
		LinkColor:             "#60A5FA", // Hyperlink color
		ToastSuccessTextColor: "#000000", // Toast success foreground
		ToastErrorTextColor:   "#FFFFFF", // Toast error foreground

		// Danger button colors
		DangerLight:  "#FCA5A5", // Light red text
		DangerDark:   "#7F1D1D", // Dark red background
		DangerBright: "#DC2626", // Bright red focused bg
		DangerHover:  "#B91C1C", // Darker red hover bg
		TextInverse:  "#FFFFFF", // Inverse/contrast text

		// Scrollbar colors
		ScrollbarTrackColor: "#4B5563", // Same as TextSubtle
		ScrollbarThumbColor: "#6B7280", // Same as TextMuted

		// Blame age gradient colors
		BlameAge1: "#34D399", // < 1 week
		BlameAge2: "#84CC16", // < 1 month
		BlameAge3: "#FBBF24", // < 3 months
		BlameAge4: "#F97316", // < 6 months
		BlameAge5: "#9CA3AF", // < 1 year

		// Third-party theme names
		SyntaxTheme:   "monokai",
		MarkdownTheme: "dark",
	}
}
