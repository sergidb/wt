package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sergidb/wt/internal/shell"
	"github.com/sergidb/wt/internal/worktree"
)

type screen int

const (
	screenList screen = iota
	screenActions
)

// Action descriptions for the actions menu
var actionDescs = map[string]string{
	"cd":     "Navigate to this worktree",
	"run":    "Start configured services",
	"info":   "Show worktree details",
	"remove": "Delete this worktree",
	"back":   "Return to worktree list",
}

type model struct {
	worktrees []worktree.Worktree
	cursor    int
	screen    screen
	actions   []string
	actionIdx int
	result    string
	quitting  bool
	repoRoot  string
}

func initialModel(repoRoot string, wts []worktree.Worktree) model {
	return model{
		worktrees: wts,
		repoRoot:  repoRoot,
		actions:   []string{"cd", "run", "info", "remove", "back"},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenList:
			return m.updateList(msg)
		case screenActions:
			return m.updateActions(msg)
		}
	}
	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.worktrees)-1 {
			m.cursor++
		}
	case "enter":
		m.screen = screenActions
		m.actionIdx = 0
	case "d":
		wt := m.worktrees[m.cursor]
		if !wt.IsMain {
			m.result = "rm:" + wt.Name
			return m, tea.Quit
		}
	case "c":
		m.result = "config"
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = screenList
	case "up", "k":
		if m.actionIdx > 0 {
			m.actionIdx--
		}
	case "down", "j":
		if m.actionIdx < len(m.actions)-1 {
			m.actionIdx++
		}
	case "enter":
		wt := m.worktrees[m.cursor]
		action := m.actions[m.actionIdx]
		switch action {
		case "cd":
			m.result = shell.CdPrefix + wt.Path
			return m, tea.Quit
		case "run":
			m.result = "run:" + wt.Path
			return m, tea.Quit
		case "info":
			m.result = "info:" + wt.Name
			return m, tea.Quit
		case "remove":
			if !wt.IsMain {
				m.result = "rm:" + wt.Name
				return m, tea.Quit
			}
		case "back":
			m.screen = screenList
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case screenList:
		return m.viewList()
	case screenActions:
		return m.viewActions()
	}
	return ""
}

func (m model) viewList() string {
	var b strings.Builder

	// Header
	b.WriteString(renderHeader("Worktrees"))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	for i, wt := range m.worktrees {
		isSelected := i == m.cursor

		// Cursor
		cursor := "   "
		if isSelected {
			cursor = cursorStyle.Render(" ▸ ")
		}

		// Name
		name := wt.Name
		if isSelected {
			name = selectedStyle.Render(name)
		} else {
			name = secondaryStyle.Render(name)
		}

		// Badge
		var badge string
		if wt.IsMain {
			badge = badgeMainStyle.Render(" ● main")
		} else if wt.Source == worktree.SourceClaude {
			badge = badgeClaudeStyle.Render(" ◆ claude")
		} else {
			badge = badgeGitStyle.Render(" ○ git")
		}

		// Branch (if different from name)
		branch := ""
		if wt.Branch != "" && wt.Branch != wt.Name {
			branch = dimStyle.Render("  " + wt.Branch)
		}

		b.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, name, badge, branch))

		// Path (always shown, relative to repo root)
		relPath := m.relativePath(wt.Path)
		if isSelected {
			b.WriteString(dimStyle.Render("     "+relPath) + "\n")
		} else {
			b.WriteString(dimStyle.Render("     "+relPath) + "\n")
		}
	}

	// Status bar
	count := fmt.Sprintf("%d worktrees", len(m.worktrees))
	pos := fmt.Sprintf("%d/%d", m.cursor+1, len(m.worktrees))
	b.WriteString(renderStatusBar(count, dimStyle.Render(pos)))

	// Separator + Help
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("↑↓ navigate  enter select  d delete  c config  q quit"))

	return b.String()
}

func (m model) viewActions() string {
	var b strings.Builder

	wt := m.worktrees[m.cursor]

	// Header with breadcrumb
	b.WriteString(renderHeader(wt.Name, "Actions"))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	// Worktree info summary
	var badge string
	if wt.IsMain {
		badge = badgeMainStyle.Render("● main")
	} else if wt.Source == worktree.SourceClaude {
		badge = badgeClaudeStyle.Render("◆ claude")
	} else {
		badge = badgeGitStyle.Render("○ git")
	}
	b.WriteString(fmt.Sprintf("   %s  %s\n", badge, dimStyle.Render(m.relativePath(wt.Path))))
	b.WriteString("\n")

	// Actions with descriptions
	for i, action := range m.actions {
		isSelected := i == m.actionIdx

		cursor := "   "
		if isSelected {
			cursor = cursorStyle.Render(" ▸ ")
		}

		name := actionNameStyle.Render(action)
		desc := actionDescStyle.Render(actionDescs[action])

		if isSelected {
			name = selectedStyle.Render(actionNameStyle.Render(action))
			desc = secondaryStyle.Render(actionDescs[action])
		}

		// Disable remove for main
		if action == "remove" && wt.IsMain {
			name = dimStyle.Render(actionNameStyle.Render(action))
			desc = dimStyle.Render("(main worktree)")
		}

		b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, name, desc))
	}

	// Separator + Help
	b.WriteString("\n")
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("↑↓ navigate  enter select  esc back  q quit"))

	return b.String()
}

func (m model) relativePath(absPath string) string {
	rel, err := filepath.Rel(m.repoRoot, absPath)
	if err != nil {
		return absPath
	}
	if rel == "." {
		return "./"
	}
	return rel
}

// Run launches the interactive TUI and returns the result string.
func Run(repoRoot string) (string, error) {
	wts, err := worktree.List(repoRoot)
	if err != nil {
		return "", err
	}

	if len(wts) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}

	m := initialModel(repoRoot, wts)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	fm := finalModel.(model)
	return fm.result, nil
}
