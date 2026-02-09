package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/sidecar/internal/modal"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/ui"
)

const (
	projectAddNameID   = "project-add-name"
	projectAddPathID   = "project-add-path"
	projectAddThemeID  = "project-add-theme"
	projectAddAddID    = "project-add-add"
	projectAddCancelID = "project-add-cancel"
)

// ensureProjectAddModal builds/rebuilds the project add modal.
func (m *Model) ensureProjectAddModal() {
	modalW := 50
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	// Only rebuild if modal doesn't exist or width changed
	if m.project.AddModal != nil && m.project.AddModalWidth == modalW {
		return
	}
	m.project.AddModalWidth = modalW

	m.project.AddModal = modal.New("Add Project",
		modal.WithWidth(modalW),
		modal.WithHints(false),
	).
		AddSection(m.projectAddNameSection()).
		AddSection(modal.Spacer()).
		AddSection(m.projectAddPathSection()).
		AddSection(modal.Spacer()).
		AddSection(m.projectAddThemeSection()).
		AddSection(modal.When(func() bool { return m.project.Add != nil && m.project.Add.errorMessage != "" }, m.projectAddErrorSection())).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(" Add ", projectAddAddID, modal.BtnPrimary()),
			modal.Btn(" Cancel ", projectAddCancelID),
		)).
		AddSection(m.projectAddHintsSection())
}

// clearProjectAddModal clears the modal state.
func (m *Model) clearProjectAddModal() {
	m.project.AddModal = nil
	m.project.AddModalWidth = 0
}

// projectAddNameSection renders the name input field.
func (m *Model) projectAddNameSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		if m.project.Add == nil {
			return modal.RenderedSection{}
		}
		var sb strings.Builder

		sb.WriteString("Name:")
		sb.WriteString("\n")

		// Sync textinput focus state with modal focus
		isFocused := focusID == projectAddNameID
		if isFocused {
			m.project.Add.nameInput.Focus()
		} else {
			m.project.Add.nameInput.Blur()
		}

		// Input field style based on focus
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.TextMuted).
			Padding(0, 1)
		if isFocused {
			inputStyle = inputStyle.BorderForeground(styles.Primary)
		}

		sb.WriteString(inputStyle.Render(m.project.Add.nameInput.View()))

		return modal.RenderedSection{
			Content: sb.String(),
			Focusables: []modal.FocusableInfo{{
				ID:      projectAddNameID,
				OffsetX: 0,
				OffsetY: 1, // After the label line
				Width:   contentWidth,
				Height:  3, // Border + content + border
			}},
		}
	}, m.projectAddNameUpdate)
}

// projectAddNameUpdate handles key events for the name input.
func (m *Model) projectAddNameUpdate(msg tea.Msg, focusID string) (string, tea.Cmd) {
	if focusID != projectAddNameID {
		return "", nil
	}
	if m.project.Add == nil {
		return "", nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(keyMsg) {
		return "", nil
	}

	// Clear error on typing
	m.project.Add.errorMessage = ""
	m.project.AddModalWidth = 0 // Force rebuild to hide error

	var cmd tea.Cmd
	m.project.Add.nameInput, cmd = m.project.Add.nameInput.Update(keyMsg)
	return "", cmd
}

// projectAddPathSection renders the path input field.
func (m *Model) projectAddPathSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		if m.project.Add == nil {
			return modal.RenderedSection{}
		}
		var sb strings.Builder

		sb.WriteString("Path:")
		sb.WriteString("\n")

		// Sync textinput focus state with modal focus
		isFocused := focusID == projectAddPathID
		if isFocused {
			m.project.Add.pathInput.Focus()
		} else {
			m.project.Add.pathInput.Blur()
		}

		// Input field style based on focus
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.TextMuted).
			Padding(0, 1)
		if isFocused {
			inputStyle = inputStyle.BorderForeground(styles.Primary)
		}

		sb.WriteString(inputStyle.Render(m.project.Add.pathInput.View()))

		return modal.RenderedSection{
			Content: sb.String(),
			Focusables: []modal.FocusableInfo{{
				ID:      projectAddPathID,
				OffsetX: 0,
				OffsetY: 1, // After the label line
				Width:   contentWidth,
				Height:  3, // Border + content + border
			}},
		}
	}, m.projectAddPathUpdate)
}

// projectAddPathUpdate handles key events for the path input.
func (m *Model) projectAddPathUpdate(msg tea.Msg, focusID string) (string, tea.Cmd) {
	if focusID != projectAddPathID {
		return "", nil
	}
	if m.project.Add == nil {
		return "", nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(keyMsg) {
		return "", nil
	}

	// Clear error on typing
	m.project.Add.errorMessage = ""
	m.project.AddModalWidth = 0 // Force rebuild to hide error

	var cmd tea.Cmd
	m.project.Add.pathInput, cmd = m.project.Add.pathInput.Update(keyMsg)
	return "", cmd
}

// projectAddThemeSection renders the theme selector field.
func (m *Model) projectAddThemeSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		var sb strings.Builder

		sb.WriteString("Theme:")
		sb.WriteString("\n")

		themeValue := "(use global)"
		if m.project.Add != nil && m.project.Add.themeSelected != "" {
			themeValue = m.project.Add.themeSelected
		}

		// Field style based on focus/hover
		fieldStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.TextMuted).
			Padding(0, 1)
		if focusID == projectAddThemeID || hoverID == projectAddThemeID {
			fieldStyle = fieldStyle.BorderForeground(styles.Primary)
		}

		sb.WriteString(fieldStyle.Render(themeValue))

		return modal.RenderedSection{
			Content: sb.String(),
			Focusables: []modal.FocusableInfo{{
				ID:      projectAddThemeID,
				OffsetX: 0,
				OffsetY: 1, // After the label line
				Width:   contentWidth,
				Height:  3, // Border + content + border
			}},
		}
	}, m.projectAddThemeUpdate)
}

// projectAddThemeUpdate handles key events for the theme field.
func (m *Model) projectAddThemeUpdate(msg tea.Msg, focusID string) (string, tea.Cmd) {
	if focusID != projectAddThemeID {
		return "", nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", nil
	}

	if keyMsg.String() == "enter" {
		return "open-theme-picker", nil
	}

	return "", nil
}

// projectAddErrorSection renders the error message.
func (m *Model) projectAddErrorSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		errStyle := lipgloss.NewStyle().Foreground(styles.Error)
		if m.project.Add == nil {
			return modal.RenderedSection{}
		}
		return modal.RenderedSection{Content: errStyle.Render(m.project.Add.errorMessage)}
	}, nil)
}

// projectAddHintsSection renders the help text.
func (m *Model) projectAddHintsSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		var sb strings.Builder

		sb.WriteString(styles.KeyHint.Render("tab"))
		sb.WriteString(styles.Muted.Render(" next  "))
		sb.WriteString(styles.KeyHint.Render("enter"))
		sb.WriteString(styles.Muted.Render(" confirm  "))
		sb.WriteString(styles.KeyHint.Render("esc"))
		sb.WriteString(styles.Muted.Render(" back"))

		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}

// renderProjectAddModal renders the project add modal using the modal library.
func (m *Model) renderProjectAddModal(content string) string {
	// If theme picker is open, render it on top
	if m.project.AddThemeMode {
		return m.renderProjectAddThemePickerOverlay(content)
	}

	m.ensureProjectAddModal()
	if m.project.AddModal == nil {
		return content
	}

	if m.project.AddMouseHandler == nil {
		m.project.AddMouseHandler = mouse.NewHandler()
	}
	modalContent := m.project.AddModal.Render(m.width, m.height, m.project.AddMouseHandler)
	return ui.OverlayModal(content, modalContent, m.width, m.height)
}

// handleProjectAddModalKeys handles keyboard input for the project add modal.
func (m *Model) handleProjectAddModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If theme picker is open, handle it separately
	if m.project.AddThemeMode {
		return m.handleProjectAddThemePickerKeys(msg)
	}

	m.ensureProjectAddModal()
	if m.project.AddModal == nil {
		return m, nil
	}

	action, cmd := m.project.AddModal.HandleKey(msg)

	switch action {
	case "cancel", projectAddCancelID:
		m.resetProjectAdd()
		return m, nil
	case "open-theme-picker":
		m.initProjectAddThemePicker()
		return m, nil
	case projectAddAddID:
		if errMsg := m.validateProjectAdd(); errMsg != "" {
			if m.project.Add != nil {
				m.project.Add.errorMessage = errMsg
			}
			m.project.AddModalWidth = 0 // Force rebuild to show error
			return m, nil
		}
		saveCmd := m.saveProjectAdd()
		m.resetProjectAdd()
		return m, saveCmd
	}

	return m, cmd
}

// handleProjectAddModalMouse handles mouse events for the project add modal.
func (m *Model) handleProjectAddModalMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// If theme picker is open, let it handle mouse
	if m.project.AddThemeMode {
		return m.handleProjectAddThemePickerMouse(msg)
	}

	m.ensureProjectAddModal()
	if m.project.AddModal == nil {
		return m, nil
	}

	if m.project.AddMouseHandler == nil {
		m.project.AddMouseHandler = mouse.NewHandler()
	}

	action := m.project.AddModal.HandleMouse(msg, m.project.AddMouseHandler)

	switch action {
	case projectAddCancelID:
		m.resetProjectAdd()
		return m, nil
	case projectAddThemeID:
		m.initProjectAddThemePicker()
		return m, nil
	case projectAddAddID:
		if errMsg := m.validateProjectAdd(); errMsg != "" {
			if m.project.Add != nil {
				m.project.Add.errorMessage = errMsg
			}
			m.project.AddModalWidth = 0 // Force rebuild to show error
			return m, nil
		}
		saveCmd := m.saveProjectAdd()
		m.resetProjectAdd()
		return m, saveCmd
	}

	return m, nil
}

// handleProjectAddThemePickerMouse handles mouse events for theme picker.
func (m *Model) handleProjectAddThemePickerMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Theme picker doesn't have dedicated mouse handling yet
	return m, nil
}
