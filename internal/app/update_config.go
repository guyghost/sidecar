package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/event"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/theme"
)

// handleConfigChanged applies a new configuration that was loaded by the
// config file watcher. It updates theme, keymap overrides, UI preferences,
// plugin context, and publishes an event for plugins.
func (m *Model) handleConfigChanged(msg ConfigChangedMsg) (tea.Model, tea.Cmd) {
	old := m.cfg
	m.cfg = msg.Config

	// 1. Re-apply theme if changed
	if old.UI.Theme.Name != msg.Config.UI.Theme.Name ||
		old.UI.Theme.Community != msg.Config.UI.Theme.Community ||
		!mapsEqualStringAny(old.UI.Theme.Overrides, msg.Config.UI.Theme.Overrides) {
		resolved := theme.ResolveTheme(msg.Config, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
	}

	// 2. Update UI preferences
	m.showClock = msg.Config.UI.ShowClock
	styles.PillTabsEnabled = msg.Config.UI.NerdFontsEnabled

	// 3. Re-apply keymap overrides if changed
	if !mapsEqualStringString(old.Keymap.Overrides, msg.Config.Keymap.Overrides) {
		m.keymap.ClearUserOverrides()
		for key, cmdID := range msg.Config.Keymap.Overrides {
			m.keymap.SetUserOverride(key, cmdID)
		}
	}

	// 4. Update plugin context config pointer so plugins see new config
	m.registry.UpdateConfig(msg.Config)

	// 5. Publish config change event on the event bus
	if bus := m.registry.EventBus(); bus != nil {
		bus.Publish(event.TopicConfigChange, event.NewEvent(
			event.TypeConfigChanged,
			event.TopicConfigChange,
			msg.Config,
		))
	}

	// 6. Toast notification
	return *m, func() tea.Msg {
		return ToastMsg{Message: "Config reloaded", Duration: 2 * time.Second}
	}
}

// mapsEqualStringString returns true if two string→string maps have the same entries.
func mapsEqualStringString(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// mapsEqualStringAny returns true if two string→any maps are shallowly equal.
// Uses JSON-style comparison for values (only checks string/number/bool).
func mapsEqualStringAny(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		// Simple equality check (works for strings, numbers, bools)
		if v != bv {
			return false
		}
	}
	return true
}
