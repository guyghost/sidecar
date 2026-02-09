package styles

import (
	"regexp"
	"sort"
	"sync"
)

// themeMu protects access to themeRegistry and currentTheme for thread safety
var themeMu sync.RWMutex

// hexColorRegex validates hex color codes (#RRGGBB or #RRGGBBAA with alpha)
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?$`)

// ColorPalette holds all theme colors (from config files)
type ColorPalette struct {
	// Brand colors
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Accent    string `json:"accent"`

	// Status colors
	Success string `json:"success"`
	Warning string `json:"warning"`
	Error   string `json:"error"`
	Info    string `json:"info"`

	// Text colors
	TextPrimary   string `json:"textPrimary"`
	TextSecondary string `json:"textSecondary"`
	TextMuted     string `json:"textMuted"`
	TextSubtle    string `json:"textSubtle"`
	TextSelection string `json:"textSelection"` // Text on selection backgrounds (BgTertiary)

	// Background colors
	BgPrimary   string `json:"bgPrimary"`
	BgSecondary string `json:"bgSecondary"`
	BgTertiary  string `json:"bgTertiary"`
	BgOverlay   string `json:"bgOverlay"`

	// Border colors
	BorderNormal string `json:"borderNormal"`
	BorderActive string `json:"borderActive"`
	BorderMuted  string `json:"borderMuted"`

	// Gradient border colors (for angled gradient borders on panels)
	GradientBorderActive []string `json:"gradientBorderActive"` // Colors for active panel gradient
	GradientBorderNormal []string `json:"gradientBorderNormal"` // Colors for inactive panel gradient
	GradientBorderAngle  float64  `json:"gradientBorderAngle"`  // Angle in degrees (default: 30)

	// Tab theme configuration
	TabStyle  string   `json:"tabStyle"`  // "gradient", "per-tab", "solid", "minimal", or preset name
	TabColors []string `json:"tabColors"` // Color stops for gradient OR per-tab colors

	// Diff colors
	DiffAddFg    string `json:"diffAddFg"`
	DiffAddBg    string `json:"diffAddBg"`
	DiffRemoveFg string `json:"diffRemoveFg"`
	DiffRemoveBg string `json:"diffRemoveBg"`

	// Additional UI colors
	TextHighlight    string `json:"textHighlight"`    // For subtitle, special text
	ButtonHover      string `json:"buttonHover"`      // Button hover state
	TabTextInactive  string `json:"tabTextInactive"`  // Inactive tab text
	Link             string `json:"link"`             // Hyperlink color
	ToastSuccessText string `json:"toastSuccessText"` // Toast success foreground
	ToastErrorText   string `json:"toastErrorText"`   // Toast error foreground

	// Danger button colors (for destructive action buttons)
	DangerLight  string `json:"dangerLight"`  // Light red for danger button text
	DangerDark   string `json:"dangerDark"`   // Dark red for danger button background
	DangerBright string `json:"dangerBright"` // Bright red for focused danger button bg
	DangerHover  string `json:"dangerHover"`  // Darker red for hover danger button bg
	TextInverse  string `json:"textInverse"`  // Inverse text (white on dark themes)

	// Scrollbar colors
	ScrollbarTrack string `json:"scrollbarTrack"` // Track color (default: TextSubtle)
	ScrollbarThumb string `json:"scrollbarThumb"` // Thumb color (default: TextMuted)

	// Blame age gradient colors (newest → oldest)
	BlameAge1 string `json:"blameAge1"` // < 1 week (light green)
	BlameAge2 string `json:"blameAge2"` // < 1 month (lime)
	BlameAge3 string `json:"blameAge3"` // < 3 months (amber)
	BlameAge4 string `json:"blameAge4"` // < 6 months (orange)
	BlameAge5 string `json:"blameAge5"` // < 1 year (gray)

	// Third-party theme names
	SyntaxTheme   string `json:"syntaxTheme"`   // Chroma theme name
	MarkdownTheme string `json:"markdownTheme"` // Glamour theme name
}

// Theme represents a complete theme configuration
type Theme struct {
	Name        string       `json:"name"`
	DisplayName string       `json:"displayName"`
	Colors      ColorPalette `json:"colors"`
}

// Built-in themes
var (
	// DefaultTheme is the current dark theme (backwards compatible)
	DefaultTheme = Theme{
		Name:        "default",
		DisplayName: "Default Dark",
		Colors: ColorPalette{
			// Brand colors
			Primary:   "#7C3AED", // Purple
			Secondary: "#3B82F6", // Blue
			Accent:    "#F59E0B", // Amber

			// Status colors
			Success: "#10B981", // Green
			Warning: "#F59E0B", // Amber
			Error:   "#EF4444", // Red
			Info:    "#3B82F6", // Blue

			// Text colors
			TextPrimary:   "#F9FAFB",
			TextSecondary: "#9CA3AF",
			TextMuted:     "#6B7280",
			TextSubtle:    "#4B5563",
			TextSelection: "#F9FAFB", // Same as TextPrimary for built-in themes

			// Background colors
			BgPrimary:   "#111827",
			BgSecondary: "#1F2937",
			BgTertiary:  "#374151",
			BgOverlay:   "#00000080",

			// Border colors
			BorderNormal: "#374151",
			BorderActive: "#7C3AED",
			BorderMuted:  "#1F2937",

			// Gradient border colors (purple → blue, 30° angle)
			GradientBorderActive: []string{"#7C3AED", "#3B82F6"},
			GradientBorderNormal: []string{"#374151", "#2D3748"},
			GradientBorderAngle:  30.0,

			// Tab theme (rainbow gradient across all tabs)
			TabStyle:  "rainbow",
			TabColors: []string{"#DC3C3C", "#3CDC3C", "#3C3CDC", "#9C3CDC"},

			// Diff colors
			DiffAddFg:    "#10B981",
			DiffAddBg:    "#0D2818",
			DiffRemoveFg: "#EF4444",
			DiffRemoveBg: "#2D1A1A",

			// Additional UI colors
			TextHighlight:    "#E5E7EB",
			ButtonHover:      "#9D174D",
			TabTextInactive:  "#1a1a1a",
			Link:             "#60A5FA", // Light blue for links
			ToastSuccessText: "#000000", // Black on green
			ToastErrorText:   "#FFFFFF", // White on red

			// Danger button colors
			DangerLight:  "#FCA5A5",
			DangerDark:   "#7F1D1D",
			DangerBright: "#DC2626",
			DangerHover:  "#B91C1C",
			TextInverse:  "#FFFFFF",

			// Blame age gradient
			BlameAge1: "#34D399",
			BlameAge2: "#84CC16",
			BlameAge3: "#FBBF24",
			BlameAge4: "#F97316",
			BlameAge5: "#9CA3AF",

			// Third-party themes
			SyntaxTheme:   "monokai",
			MarkdownTheme: "dark",
		},
	}

	// DraculaTheme is a Dracula-inspired dark theme with vibrant colors
	DraculaTheme = Theme{
		Name:        "dracula",
		DisplayName: "Dracula",
		Colors: ColorPalette{
			// Brand colors - Dracula palette
			Primary:   "#BD93F9", // Purple
			Secondary: "#8BE9FD", // Cyan
			Accent:    "#FFB86C", // Orange

			// Status colors
			Success: "#50FA7B", // Green
			Warning: "#FFB86C", // Orange
			Error:   "#FF5555", // Red
			Info:    "#8BE9FD", // Cyan

			// Text colors
			TextPrimary:   "#F8F8F2", // Foreground
			TextSecondary: "#BFBFBF",
			TextMuted:     "#6272A4", // Comment
			TextSubtle:    "#44475A", // Current Line
			TextSelection: "#F8F8F2", // Same as TextPrimary for built-in themes

			// Background colors
			BgPrimary:   "#282A36", // Background
			BgSecondary: "#343746",
			BgTertiary:  "#44475A", // Current Line
			BgOverlay:   "#00000080",

			// Border colors
			BorderNormal: "#44475A",
			BorderActive: "#BD93F9",
			BorderMuted:  "#343746",

			// Gradient border colors (purple → cyan, 30° angle)
			GradientBorderActive: []string{"#BD93F9", "#8BE9FD"},
			GradientBorderNormal: []string{"#44475A", "#383A4A"},
			GradientBorderAngle:  30.0,

			// Tab theme (Dracula purple-pink-cyan gradient)
			TabStyle:  "gradient",
			TabColors: []string{"#BD93F9", "#FF79C6", "#8BE9FD"},

			// Diff colors
			DiffAddFg:    "#50FA7B",
			DiffAddBg:    "#1E3A29",
			DiffRemoveFg: "#FF5555",
			DiffRemoveBg: "#3D2A2A",

			// Additional UI colors
			TextHighlight:    "#F8F8F2",
			ButtonHover:      "#FF79C6", // Pink
			TabTextInactive:  "#282A36",
			Link:             "#8BE9FD", // Cyan for links (Dracula)
			ToastSuccessText: "#282A36", // Dark bg on green
			ToastErrorText:   "#F8F8F2", // Light on red

			// Danger button colors
			DangerLight:  "#FFADAD",
			DangerDark:   "#3D1F1F",
			DangerBright: "#FF5555",
			DangerHover:  "#E63E3E",
			TextInverse:  "#F8F8F2",

			// Blame age gradient
			BlameAge1: "#69FF94",
			BlameAge2: "#A4E22E",
			BlameAge3: "#FFB86C",
			BlameAge4: "#FF7979",
			BlameAge5: "#6272A4",

			// Third-party themes
			SyntaxTheme:   "dracula",
			MarkdownTheme: "dark",
		},
	}

	// MolokaiTheme is a vibrant, high-contrast theme
	MolokaiTheme = Theme{
		Name:        "molokai",
		DisplayName: "Molokai",
		Colors: ColorPalette{
			Primary:   "#F92672", // Pink
			Secondary: "#66D9EF", // Blue
			Accent:    "#A6E22E", // Green

			Success: "#A6E22E", // Green
			Warning: "#FD971F", // Orange
			Error:   "#F92672", // Red
			Info:    "#66D9EF", // Blue

			TextPrimary:   "#F8F8F2",
			TextSecondary: "#CFD0C2",
			TextMuted:     "#75715E",
			TextSubtle:    "#465457",
			TextSelection: "#F8F8F2", // Same as TextPrimary for built-in themes

			BgPrimary:   "#1B1D1E",
			BgSecondary: "#272822",
			BgTertiary:  "#3E3D32",
			BgOverlay:   "#00000080",

			BorderNormal: "#465457",
			BorderActive: "#F92672",
			BorderMuted:  "#3E3D32",

			GradientBorderActive: []string{"#F92672", "#A6E22E"},
			GradientBorderNormal: []string{"#465457", "#3E3D32"},
			GradientBorderAngle:  45.0,

			TabStyle:  "solid",
			TabColors: []string{"#F92672"},

			DiffAddFg:    "#A6E22E",
			DiffAddBg:    "#13210C",
			DiffRemoveFg: "#F92672",
			DiffRemoveBg: "#210C11",

			TextHighlight:    "#E6DB74", // Yellow
			ButtonHover:      "#F92672",
			TabTextInactive:  "#75715E",
			Link:             "#66D9EF",
			ToastSuccessText: "#1B1D1E",
			ToastErrorText:   "#F8F8F2",

			// Danger button colors
			DangerLight:  "#F8A0B8",
			DangerDark:   "#3D0F1E",
			DangerBright: "#F92672",
			DangerHover:  "#D91E63",
			TextInverse:  "#F8F8F2",

			// Blame age gradient
			BlameAge1: "#A6E22E",
			BlameAge2: "#E6DB74",
			BlameAge3: "#FD971F",
			BlameAge4: "#F92672",
			BlameAge5: "#75715E",

			SyntaxTheme:   "monokai",
			MarkdownTheme: "dark",
		},
	}

	// NordTheme is an arctic, north-bluish color palette
	NordTheme = Theme{
		Name:        "nord",
		DisplayName: "Nord",
		Colors: ColorPalette{
			Primary:   "#88C0D0", // Frost Cyan
			Secondary: "#81A1C1", // Frost Blue
			Accent:    "#EBCB8B", // Aurora Yellow

			Success: "#A3BE8C", // Aurora Green
			Warning: "#EBCB8B", // Aurora Yellow
			Error:   "#BF616A", // Aurora Red
			Info:    "#88C0D0", // Frost Cyan

			TextPrimary:   "#D8DEE9", // Snow Storm 1
			TextSecondary: "#E5E9F0", // Snow Storm 2
			TextMuted:     "#4C566A", // Polar Night 4
			TextSubtle:    "#434C5E", // Polar Night 3
			TextSelection: "#D8DEE9", // Same as TextPrimary for built-in themes

			BgPrimary:   "#2E3440", // Polar Night 1
			BgSecondary: "#3B4252", // Polar Night 2
			BgTertiary:  "#434C5E", // Polar Night 3
			BgOverlay:   "#2E3440CC",

			BorderNormal: "#4C566A",
			BorderActive: "#88C0D0",
			BorderMuted:  "#3B4252",

			GradientBorderActive: []string{"#88C0D0", "#81A1C1"},
			GradientBorderNormal: []string{"#434C5E", "#3B4252"},
			GradientBorderAngle:  120.0,

			TabStyle:  "minimal",
			TabColors: []string{"#88C0D0"},

			DiffAddFg:    "#A3BE8C",
			DiffAddBg:    "#233129",
			DiffRemoveFg: "#BF616A",
			DiffRemoveBg: "#312325",

			TextHighlight:    "#ECEFF4",
			ButtonHover:      "#5E81AC", // Frost Dark Blue
			TabTextInactive:  "#4C566A",
			Link:             "#88C0D0",
			ToastSuccessText: "#2E3440",
			ToastErrorText:   "#E5E9F0",

			// Danger button colors
			DangerLight:  "#D08770",
			DangerDark:   "#3B2A25",
			DangerBright: "#BF616A",
			DangerHover:  "#A5545C",
			TextInverse:  "#ECEFF4",

			// Blame age gradient
			BlameAge1: "#A3BE8C",
			BlameAge2: "#EBCB8B",
			BlameAge3: "#D08770",
			BlameAge4: "#BF616A",
			BlameAge5: "#4C566A",

			SyntaxTheme:   "nord",
			MarkdownTheme: "dark",
		},
	}

	// SolarizedDarkTheme is a precision color scheme
	SolarizedDarkTheme = Theme{
		Name:        "solarized-dark",
		DisplayName: "Solarized Dark",
		Colors: ColorPalette{
			Primary:   "#268BD2", // Blue
			Secondary: "#2AA198", // Cyan
			Accent:    "#B58900", // Yellow

			Success: "#859900", // Green
			Warning: "#B58900", // Yellow
			Error:   "#DC322F", // Red
			Info:    "#268BD2", // Blue

			TextPrimary:   "#93A1A1", // Base1
			TextSecondary: "#839496", // Base0
			TextMuted:     "#586E75", // Base01
			TextSubtle:    "#073642", // Base02
			TextSelection: "#93A1A1", // Same as TextPrimary for built-in themes

			BgPrimary:   "#002B36", // Base03
			BgSecondary: "#073642", // Base02
			BgTertiary:  "#002B36", // Base03 (Repeat for depth)
			BgOverlay:   "#00181ECC",

			BorderNormal: "#586E75",
			BorderActive: "#268BD2",
			BorderMuted:  "#073642",

			GradientBorderActive: []string{"#268BD2", "#2AA198"},
			GradientBorderNormal: []string{"#586E75", "#073642"},
			GradientBorderAngle:  90.0,

			TabStyle:  "solid",
			TabColors: []string{"#2AA198"},

			DiffAddFg:    "#859900",
			DiffAddBg:    "#002B36",
			DiffRemoveFg: "#DC322F",
			DiffRemoveBg: "#002B36",

			TextHighlight:    "#FDF6E3", // Base3
			ButtonHover:      "#CB4B16", // Orange
			TabTextInactive:  "#586E75",
			Link:             "#268BD2",
			ToastSuccessText: "#FDF6E3",
			ToastErrorText:   "#FDF6E3",

			// Danger button colors
			DangerLight:  "#E8A0A0",
			DangerDark:   "#2A1515",
			DangerBright: "#DC322F",
			DangerHover:  "#C12926",
			TextInverse:  "#FDF6E3",

			// Blame age gradient
			BlameAge1: "#859900",
			BlameAge2: "#B58900",
			BlameAge3: "#CB4B16",
			BlameAge4: "#DC322F",
			BlameAge5: "#586E75",

			SyntaxTheme:   "solarized-dark",
			MarkdownTheme: "dark",
		},
	}

	// TokyoNightTheme is a clean, dark theme that celebrates the lights of Downtown Tokyo
	TokyoNightTheme = Theme{
		Name:        "tokyo-night",
		DisplayName: "Tokyo Night",
		Colors: ColorPalette{
			Primary:   "#7AA2F7", // Blue
			Secondary: "#BB9AF7", // Purple
			Accent:    "#FF9E64", // Orange

			Success: "#9ECE6A", // Green
			Warning: "#E0AF68", // Yellow
			Error:   "#F7768E", // Red
			Info:    "#7DCFFF", // Cyan

			TextPrimary:   "#C0CAF5",
			TextSecondary: "#A9B1D6",
			TextMuted:     "#565F89",
			TextSubtle:    "#414868",
			TextSelection: "#C0CAF5", // Same as TextPrimary for built-in themes

			BgPrimary:   "#1A1B26",
			BgSecondary: "#24283B",
			BgTertiary:  "#414868",
			BgOverlay:   "#15161ECC",

			BorderNormal: "#565F89",
			BorderActive: "#7AA2F7",
			BorderMuted:  "#24283B",

			GradientBorderActive: []string{"#7AA2F7", "#BB9AF7"},
			GradientBorderNormal: []string{"#565F89", "#414868"},
			GradientBorderAngle:  60.0,

			TabStyle:  "gradient",
			TabColors: []string{"#7AA2F7", "#BB9AF7", "#F7768E"},

			DiffAddFg:    "#9ECE6A",
			DiffAddBg:    "#283B4D",
			DiffRemoveFg: "#F7768E",
			DiffRemoveBg: "#3F2D3D",

			TextHighlight:    "#C0CAF5",
			ButtonHover:      "#BB9AF7",
			TabTextInactive:  "#565F89",
			Link:             "#73DACA",
			ToastSuccessText: "#15161E",
			ToastErrorText:   "#C0CAF5",

			// Danger button colors
			DangerLight:  "#F7A8B8",
			DangerDark:   "#2D1520",
			DangerBright: "#F7768E",
			DangerHover:  "#E05F77",
			TextInverse:  "#C0CAF5",

			// Blame age gradient
			BlameAge1: "#9ECE6A",
			BlameAge2: "#E0AF68",
			BlameAge3: "#FF9E64",
			BlameAge4: "#F7768E",
			BlameAge5: "#565F89",

			SyntaxTheme:   "tokyo-night",
			MarkdownTheme: "dark",
		},
	}
)

// themeRegistry holds all available themes
var themeRegistry = map[string]Theme{
	"default":        DefaultTheme,
	"dracula":        DraculaTheme,
	"molokai":        MolokaiTheme,
	"nord":           NordTheme,
	"solarized-dark": SolarizedDarkTheme,
	"tokyo-night":    TokyoNightTheme,
}

// currentTheme tracks the active theme name
var currentTheme = "default"
var currentThemeValue = DefaultTheme

// IsValidHexColor checks if a string is a valid hex color code (#RRGGBB or #RRGGBBAA)
func IsValidHexColor(hex string) bool {
	return hexColorRegex.MatchString(hex)
}

// IsValidTheme checks if a theme name exists in the registry
func IsValidTheme(name string) bool {
	themeMu.RLock()
	defer themeMu.RUnlock()
	_, ok := themeRegistry[name]
	return ok
}

// GetTheme returns a theme by name, or the default theme if not found
func GetTheme(name string) Theme {
	themeMu.RLock()
	defer themeMu.RUnlock()
	if theme, ok := themeRegistry[name]; ok {
		return theme
	}
	return DefaultTheme
}

// GetCurrentTheme returns the currently active theme
func GetCurrentTheme() Theme {
	themeMu.RLock()
	theme := currentThemeValue
	themeMu.RUnlock()
	return theme
}

// GetCurrentThemeName returns the name of the currently active theme
func GetCurrentThemeName() string {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return currentTheme
}

// ListThemes returns the names of all available themes in sorted order
func ListThemes() []string {
	themeMu.RLock()
	defer themeMu.RUnlock()
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterTheme adds a custom theme to the registry
func RegisterTheme(theme Theme) {
	themeMu.Lock()
	defer themeMu.Unlock()
	themeRegistry[theme.Name] = theme
}

// paletteFromColorPalette converts a ColorPalette (from Theme) to a Palette (for NewStyles)
func paletteFromColorPalette(c ColorPalette) Palette {
	return Palette{
		Primary:   c.Primary,
		Secondary: c.Secondary,
		Accent:    c.Accent,

		Success: c.Success,
		Warning: c.Warning,
		Error:   c.Error,
		Info:    c.Info,

		TextPrimary:        c.TextPrimary,
		TextSecondary:      c.TextSecondary,
		TextMuted:          c.TextMuted,
		TextSubtle:         c.TextSubtle,
		TextSelectionColor: c.TextSelection,

		BgPrimary:   c.BgPrimary,
		BgSecondary: c.BgSecondary,
		BgTertiary:  c.BgTertiary,
		BgOverlay:   c.BgOverlay,

		BorderNormal: c.BorderNormal,
		BorderActive: c.BorderActive,
		BorderMuted:  c.BorderMuted,

		DiffAddFg:    c.DiffAddFg,
		DiffRemoveFg: c.DiffRemoveFg,
		DiffAddBg:    c.DiffAddBg,
		DiffRemoveBg: c.DiffRemoveBg,

		TextHighlight:         c.TextHighlight,
		ButtonHoverColor:      c.ButtonHover,
		TabTextInactiveColor:  c.TabTextInactive,
		LinkColor:             c.Link,
		ToastSuccessTextColor: c.ToastSuccessText,
		ToastErrorTextColor:   c.ToastErrorText,

		DangerLight:  c.DangerLight,
		DangerDark:   c.DangerDark,
		DangerBright: c.DangerBright,
		DangerHover:  c.DangerHover,
		TextInverse:  c.TextInverse,

		ScrollbarTrackColor: c.ScrollbarTrack,
		ScrollbarThumbColor: c.ScrollbarThumb,

		BlameAge1: c.BlameAge1,
		BlameAge2: c.BlameAge2,
		BlameAge3: c.BlameAge3,
		BlameAge4: c.BlameAge4,
		BlameAge5: c.BlameAge5,

		SyntaxTheme:   c.SyntaxTheme,
		MarkdownTheme: c.MarkdownTheme,
	}
}

// ApplyTheme applies a theme by name, creating a new immutable Styles instance
// and atomically swapping it into Current. This is thread-safe.
func ApplyTheme(name string) {
	theme := GetTheme(name)
	ApplyThemeColors(theme)
	themeMu.Lock()
	currentTheme = name
	currentThemeValue = theme
	themeMu.Unlock()
}

// ApplyThemeColors creates a new Styles instance from a Theme and atomically
// swaps it into Current. This is thread-safe.
func ApplyThemeColors(theme Theme) {
	c := theme.Colors

	// Update tab theme state (mutable state separate from immutable styles)
	CurrentTabStyle = c.TabStyle
	CurrentTabColors = parseTabColors(c.TabColors)

	// Update third-party theme names
	CurrentSyntaxTheme = c.SyntaxTheme
	CurrentMarkdownTheme = c.MarkdownTheme

	// Build palette and create new Styles instance
	palette := paletteFromColorPalette(c)
	newStyles := NewStyles(palette)

	// Atomically swap current styles
	Current.Store(newStyles)

	// Sync to package-level variables for backward compatibility
	syncFromStyles(newStyles)

	themeMu.Lock()
	currentThemeValue = theme
	themeMu.Unlock()
}

// ApplyThemeWithOverrides applies a theme with color overrides from config.
// This creates a new immutable Styles instance and atomically swaps it.
func ApplyThemeWithOverrides(name string, overrides map[string]string) {
	theme := GetTheme(name)

	// Apply overrides to the color palette
	if overrides != nil {
		applyOverrides(&theme.Colors, overrides)
	}

	ApplyThemeColors(theme)
	themeMu.Lock()
	currentTheme = name
	themeMu.Unlock()
}

// applyOverrides applies color overrides to a palette.
// Delegates to applySingleOverride which validates hex colors.
func applyOverrides(palette *ColorPalette, overrides map[string]string) {
	for key, value := range overrides {
		applySingleOverride(palette, key, value)
	}
}

// ApplyThemeWithGenericOverrides applies a theme with overrides that may include arrays.
// This supports gradient array overrides from YAML config.
func ApplyThemeWithGenericOverrides(name string, overrides map[string]interface{}) {
	theme := GetTheme(name)

	// Apply overrides to the color palette
	if overrides != nil {
		applyGenericOverrides(&theme.Colors, overrides)
	}

	ApplyThemeColors(theme)
	themeMu.Lock()
	currentTheme = name
	themeMu.Unlock()
}

// applyGenericOverrides applies overrides that may include arrays (for gradients).
func applyGenericOverrides(palette *ColorPalette, overrides map[string]interface{}) {
	for key, value := range overrides {
		switch v := value.(type) {
		case string:
			applySingleOverride(palette, key, v)
		case []interface{}:
			// Handle array values (for gradient colors)
			colors := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					colors = append(colors, s)
				}
			}
			applyArrayOverride(palette, key, colors)
		case []string:
			applyArrayOverride(palette, key, v)
		case float64:
			applyFloatOverride(palette, key, v)
		case int:
			applyFloatOverride(palette, key, float64(v))
		}
	}
}

// applySingleOverride applies a single string override.
// Color values must be valid hex colors (#RRGGBB). Invalid colors are silently ignored.
func applySingleOverride(palette *ColorPalette, key, value string) {
	// syntaxTheme, markdownTheme, and tabStyle are names, not colors
	isThemeName := key == "syntaxTheme" || key == "markdownTheme" || key == "tabStyle"
	if !isThemeName && !IsValidHexColor(value) {
		return // Skip invalid hex color
	}

	switch key {
	case "primary":
		palette.Primary = value
	case "secondary":
		palette.Secondary = value
	case "accent":
		palette.Accent = value
	case "success":
		palette.Success = value
	case "warning":
		palette.Warning = value
	case "error":
		palette.Error = value
	case "info":
		palette.Info = value
	case "textPrimary":
		palette.TextPrimary = value
	case "textSecondary":
		palette.TextSecondary = value
	case "textMuted":
		palette.TextMuted = value
	case "textSubtle":
		palette.TextSubtle = value
	case "textSelection":
		palette.TextSelection = value
	case "bgPrimary":
		palette.BgPrimary = value
	case "bgSecondary":
		palette.BgSecondary = value
	case "bgTertiary":
		palette.BgTertiary = value
	case "bgOverlay":
		palette.BgOverlay = value
	case "borderNormal":
		palette.BorderNormal = value
	case "borderActive":
		palette.BorderActive = value
	case "borderMuted":
		palette.BorderMuted = value
	case "diffAddFg":
		palette.DiffAddFg = value
	case "diffAddBg":
		palette.DiffAddBg = value
	case "diffRemoveFg":
		palette.DiffRemoveFg = value
	case "diffRemoveBg":
		palette.DiffRemoveBg = value
	case "textHighlight":
		palette.TextHighlight = value
	case "buttonHover":
		palette.ButtonHover = value
	case "tabTextInactive":
		palette.TabTextInactive = value
	case "link":
		palette.Link = value
	case "toastSuccessText":
		palette.ToastSuccessText = value
	case "toastErrorText":
		palette.ToastErrorText = value
	case "syntaxTheme":
		palette.SyntaxTheme = value
	case "markdownTheme":
		palette.MarkdownTheme = value
	case "tabStyle":
		palette.TabStyle = value
	case "dangerLight":
		palette.DangerLight = value
	case "dangerDark":
		palette.DangerDark = value
	case "dangerBright":
		palette.DangerBright = value
	case "dangerHover":
		palette.DangerHover = value
	case "textInverse":
		palette.TextInverse = value
	case "blameAge1":
		palette.BlameAge1 = value
	case "blameAge2":
		palette.BlameAge2 = value
	case "blameAge3":
		palette.BlameAge3 = value
	case "blameAge4":
		palette.BlameAge4 = value
	case "blameAge5":
		palette.BlameAge5 = value
	case "scrollbarTrack":
		palette.ScrollbarTrack = value
	case "scrollbarThumb":
		palette.ScrollbarThumb = value
	}
}

// applyArrayOverride applies an array override (for gradient colors).
// All colors must be valid hex colors. The entire array is rejected if any color is invalid.
func applyArrayOverride(palette *ColorPalette, key string, colors []string) {
	// Validate all colors in the array
	for _, c := range colors {
		if !IsValidHexColor(c) {
			return // Reject entire array if any color is invalid
		}
	}

	switch key {
	case "gradientBorderActive":
		palette.GradientBorderActive = colors
	case "gradientBorderNormal":
		palette.GradientBorderNormal = colors
	case "tabColors":
		palette.TabColors = colors
	}
}

// applyFloatOverride applies a float override (for gradient angle).
func applyFloatOverride(palette *ColorPalette, key string, value float64) {
	switch key {
	case "gradientBorderAngle":
		palette.GradientBorderAngle = value
	}
}

// GetSyntaxTheme returns current syntax highlighting theme name
func GetSyntaxTheme() string {
	return CurrentSyntaxTheme
}

// GetMarkdownTheme returns current markdown rendering theme name
func GetMarkdownTheme() string {
	return CurrentMarkdownTheme
}
