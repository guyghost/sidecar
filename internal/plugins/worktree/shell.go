package worktree

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Shell session constants
const (
	shellSessionPrefix = "sidecar-sh-" // Distinct from worktree prefix "sidecar-wt-"
)

// Shell session messages
type (
	// ShellCreatedMsg signals shell session was created
	ShellCreatedMsg struct{}

	// ShellDetachedMsg signals user detached from shell session
	ShellDetachedMsg struct {
		Err error
	}

	// ShellKilledMsg signals shell session was terminated
	ShellKilledMsg struct{}

	// ShellOutputMsg signals shell output was captured (for polling)
	ShellOutputMsg struct {
		Output  string
		Changed bool
	}
)

// initShellSession initializes shell session tracking for the current project.
// Called from Init() to check for existing sessions from previous runs.
func (p *Plugin) initShellSession() {
	projectName := filepath.Base(p.ctx.WorkDir)
	p.shellSessionName = shellSessionPrefix + sanitizeName(projectName)

	// Check if session already exists from previous run
	if sessionExists(p.shellSessionName) {
		p.shellSession = &Agent{
			Type:        AgentShell,
			TmuxSession: p.shellSessionName,
			OutputBuf:   NewOutputBuffer(outputBufferCap),
			StartedAt:   time.Now(), // Approximate, we don't know actual start
			Status:      AgentStatusRunning,
		}
	}
}

// createShellSession creates a new tmux session in the project root directory.
func (p *Plugin) createShellSession() tea.Cmd {
	return func() tea.Msg {
		// Check if session already exists
		if sessionExists(p.shellSessionName) {
			// Session exists, just update our state
			p.shellSession = &Agent{
				Type:        AgentShell,
				TmuxSession: p.shellSessionName,
				OutputBuf:   NewOutputBuffer(outputBufferCap),
				StartedAt:   time.Now(),
				Status:      AgentStatusRunning,
			}
			return ShellCreatedMsg{}
		}

		// Create new detached session in project directory
		args := []string{
			"new-session",
			"-d",                    // Detached
			"-s", p.shellSessionName, // Session name
			"-c", p.ctx.WorkDir,      // Working directory
		}
		cmd := exec.Command("tmux", args...)
		if err := cmd.Run(); err != nil {
			return ShellDetachedMsg{Err: fmt.Errorf("create shell session: %w", err)}
		}

		// Track as managed session
		p.managedSessions[p.shellSessionName] = true

		// Create agent struct for tracking
		p.shellSession = &Agent{
			Type:        AgentShell,
			TmuxSession: p.shellSessionName,
			OutputBuf:   NewOutputBuffer(outputBufferCap),
			StartedAt:   time.Now(),
			Status:      AgentStatusRunning,
		}

		return ShellCreatedMsg{}
	}
}

// attachToShell attaches to the shell tmux session.
func (p *Plugin) attachToShell() tea.Cmd {
	if p.shellSession == nil || p.shellSessionName == "" {
		return nil
	}

	c := exec.Command("tmux", "attach-session", "-t", p.shellSessionName)
	projectName := filepath.Base(p.ctx.WorkDir)
	return tea.Sequence(
		tea.Printf("\nAttaching to %s shell. Press Ctrl-b d to return to sidecar.\n", projectName),
		tea.ExecProcess(c, func(err error) tea.Msg {
			return ShellDetachedMsg{Err: err}
		}),
	)
}

// killShellSession terminates the shell tmux session.
func (p *Plugin) killShellSession() tea.Cmd {
	if p.shellSessionName == "" {
		return nil
	}

	return func() tea.Msg {
		// Kill the session
		cmd := exec.Command("tmux", "kill-session", "-t", p.shellSessionName)
		cmd.Run() // Ignore errors (session may already be dead)

		// Clean up tracking
		delete(p.managedSessions, p.shellSessionName)
		globalPaneCache.remove(p.shellSessionName)

		// Clear session state
		p.shellSession = nil

		return ShellKilledMsg{}
	}
}

// pollShellSession captures output from the shell tmux session.
func (p *Plugin) pollShellSession() tea.Cmd {
	if p.shellSession == nil || p.shellSessionName == "" {
		return nil
	}

	return func() tea.Msg {
		output, err := capturePaneDirect(p.shellSessionName)
		if err != nil {
			return ShellOutputMsg{Output: "", Changed: false}
		}

		// Trim to max bytes
		output = trimCapturedOutput(output, p.tmuxCaptureMaxBytes)

		// Update buffer and check if content changed
		changed := p.shellSession.OutputBuf.Update(output)
		if changed {
			p.shellSession.LastOutput = time.Now()
		}

		return ShellOutputMsg{Output: output, Changed: changed}
	}
}

// scheduleShellPoll schedules a poll for shell output after delay.
func (p *Plugin) scheduleShellPoll(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return pollShellMsg{}
	})
}

// pollShellMsg triggers a shell output poll.
type pollShellMsg struct{}

// shellOutputVisible returns true if shell output is currently visible.
func (p *Plugin) shellOutputVisible() bool {
	return p.focused &&
		p.viewMode == ViewModeList &&
		p.shellSelected &&
		p.previewTab == PreviewTabOutput
}

// shellPollInterval returns appropriate poll interval based on visibility.
func (p *Plugin) shellPollInterval() time.Duration {
	if p.shellOutputVisible() {
		return pollIntervalActive
	}
	if p.focused {
		return pollIntervalBackground
	}
	return pollIntervalUnfocused
}
