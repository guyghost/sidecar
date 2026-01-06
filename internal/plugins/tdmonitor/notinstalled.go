package tdmonitor

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Stallion ASCII art - a galloping horse
const stallionArt = "" +
	"                 >\\/7\n" +
	"             _.-(6'  \\\n" +
	"            (=___._/` \\\n" +
	"                 )  \\ |\n" +
	"                /   / |\n" +
	"               /    > /\n" +
	"              j    < _\\\n" +
	"          _.-' :      ``.\n" +
	"          \\ r=._\\        `.\n" +
	"         <`\\\\_  \\         . `-. \n" +
	"          \\ r-7  `-. ._  ' .  `\\\n" +
	"           \\`,      `-.`7  7)   )\n" +
	"            \\/         \\|  \\'  / `-._\n" +
	"                       ||    .'\n" +
	"                        \\\\  (\n" +
	"                         >\\  >\n" +
	"                     ,.-' >.'\n" +
	"                    <.'_.''\n" +
	"                      <'\n"

// Color palette for gradient animation
var (
	colorPurple = hexToRGB("#7C3AED")
	colorBlue   = hexToRGB("#3B82F6")
	colorAmber  = hexToRGB("#F59E0B")
)

// RGB represents a color in RGB space for interpolation.
type RGB struct {
	R, G, B float64
}

// hexToRGB converts a hex color string to RGB.
func hexToRGB(hex string) RGB {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return RGB{float64(r), float64(g), float64(b)}
}

// toLipgloss converts RGB back to a lipgloss Color.
func (c RGB) toLipgloss() lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", int(c.R), int(c.G), int(c.B)))
}

// lerpRGB linearly interpolates between two colors.
func lerpRGB(c1, c2 RGB, t float64) RGB {
	return RGB{
		R: c1.R + (c2.R-c1.R)*t,
		G: c1.G + (c2.G-c1.G)*t,
		B: c1.B + (c2.B-c1.B)*t,
	}
}

// NotInstalledModel handles the animated "td not installed" view.
type NotInstalledModel struct {
	startTime time.Time
	width     int
	height    int
}

// NewNotInstalledModel creates a new not-installed view model.
func NewNotInstalledModel() *NotInstalledModel {
	return &NotInstalledModel{
		startTime: time.Now(),
	}
}

// StallionTickMsg is sent to update the animation frame.
type StallionTickMsg time.Time

// StallionTick returns a command that ticks for animation.
func StallionTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return StallionTickMsg(t)
	})
}

// Init returns the initial command (starts animation).
func (m *NotInstalledModel) Init() tea.Cmd {
	return StallionTick()
}

// Update handles messages for the not-installed view.
func (m *NotInstalledModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case StallionTickMsg:
		return StallionTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return nil
}

// gradientColorAt returns the color for a character based on its position and time.
// Creates a sweeping gradient effect across the image.
func (m *NotInstalledModel) gradientColorAt(charIndex, totalChars int) RGB {
	elapsed := time.Since(m.startTime).Seconds()
	cycleDuration := 6.0 // slower: 6 seconds per full cycle

	// Calculate wave position (0 to 1) that sweeps across the image
	wavePos := math.Mod(elapsed/cycleDuration, 1.0)

	// Character's normalized position in the art (0 to 1)
	charPos := float64(charIndex) / float64(totalChars)

	// Distance from wave center, wrapped for continuous effect
	dist := charPos - wavePos
	if dist < -0.5 {
		dist += 1.0
	} else if dist > 0.5 {
		dist -= 1.0
	}

	// Convert distance to color blend factor
	// Characters near the wave center get the "highlight" color
	t := math.Abs(dist) * 3.0 // Scale factor for gradient width
	if t > 1.0 {
		t = 1.0
	}

	// Three-color gradient based on wave position
	phase := math.Mod(wavePos*3.0, 1.0)

	var baseColor, highlightColor RGB
	if phase < 0.33 {
		baseColor = colorPurple
		highlightColor = colorBlue
	} else if phase < 0.66 {
		baseColor = colorBlue
		highlightColor = colorAmber
	} else {
		baseColor = colorAmber
		highlightColor = colorPurple
	}

	// Blend between highlight (at wave center) and base color
	return lerpRGB(highlightColor, baseColor, t)
}

// renderStallion returns the stallion art with animated gradient sweep.
func (m *NotInstalledModel) renderStallion() string {
	lines := strings.Split(stallionArt, "\n")

	// Count total visible characters for position calculation
	var totalChars int
	for _, line := range lines {
		for _, ch := range line {
			if ch != ' ' && ch != '\t' {
				totalChars++
			}
		}
	}

	// Render each character with its gradient color
	var result strings.Builder
	charIndex := 0

	for _, line := range lines {
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				result.WriteRune(ch)
			} else {
				color := m.gradientColorAt(charIndex, totalChars)
				style := lipgloss.NewStyle().Foreground(color.toLipgloss())
				result.WriteString(style.Render(string(ch)))
				charIndex++
			}
		}
		result.WriteRune('\n')
	}

	return result.String()
}

// renderPitch returns the professional pitch copy.
func (m *NotInstalledModel) renderPitch() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E5E7EB"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB"))

	linkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Underline(true)

	codeBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(0, 1)

	// Build content
	var b strings.Builder

	// Explain why they're seeing this screen
	b.WriteString(mutedStyle.Render("td is not installed in this project."))
	b.WriteString("\n\n")

	b.WriteString(titleStyle.Render("External memory for AI sessions"))
	b.WriteString("\n\n")

	b.WriteString(textStyle.Render("td gives each session:"))
	b.WriteString("\n")
	b.WriteString(bulletStyle.Render("  • Current focus and pending work"))
	b.WriteString("\n")
	b.WriteString(bulletStyle.Render("  • Decisions and their reasoning"))
	b.WriteString("\n")
	b.WriteString(bulletStyle.Render("  • Structured handoffs between sessions"))
	b.WriteString("\n\n")

	b.WriteString(mutedStyle.Render("Local SQLite. No cloud. Git-friendly."))
	b.WriteString("\n\n")

	// GitHub link
	b.WriteString(textStyle.Render("Learn more: "))
	b.WriteString(linkStyle.Render("https://github.com/marcus/td"))
	b.WriteString("\n\n")

	installCmd := "go install github.com/marcus/td/cmd/td@latest\ntd init"
	b.WriteString(codeBoxStyle.Render(installCmd))

	return b.String()
}

// View renders the complete not-installed screen.
func (m *NotInstalledModel) View(width, height int) string {
	m.width = width
	m.height = height

	stallion := m.renderStallion()
	pitch := m.renderPitch()

	// Combine vertically with centering
	content := lipgloss.JoinVertical(lipgloss.Center, stallion, pitch)

	// Center in available space
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
