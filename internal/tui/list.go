package tui

import (
	"fmt"
	"os"
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
		actions:   []string{"cd", "info", "remove", "back"},
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

	b.WriteString(titleStyle.Render("  Worktrees"))
	b.WriteString("\n")

	for i, wt := range m.worktrees {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		name := wt.Name
		if i == m.cursor {
			name = selectedStyle.Render(name)
		}

		var badge string
		if wt.IsMain {
			badge = mainBadgeStyle.Render(" [main]")
		} else if wt.Source == worktree.SourceClaude {
			badge = sourceClaudeStyle.Render(" [claude]")
		} else {
			badge = sourceGitStyle.Render(" [git]")
		}

		branch := ""
		if wt.Branch != "" && wt.Branch != wt.Name {
			branch = dimStyle.Render(" (" + wt.Branch + ")")
		}

		path := dimStyle.Render("  " + wt.Path)

		b.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, name, badge, branch))
		if i == m.cursor {
			b.WriteString(path + "\n")
		}
	}

	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • d delete • q quit"))

	return b.String()
}

func (m model) viewActions() string {
	var b strings.Builder

	wt := m.worktrees[m.cursor]
	b.WriteString(titleStyle.Render("  " + wt.Name))
	b.WriteString("\n")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.actionIdx {
			cursor = cursorStyle.Render("> ")
		}

		label := action
		if i == m.actionIdx {
			label = selectedStyle.Render(label)
		}

		// Disable remove for main
		if action == "remove" && wt.IsMain {
			label = dimStyle.Render(action + " (main)")
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, label))
	}

	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • esc back • q quit"))

	return b.String()
}

// Run launches the interactive TUI and returns the result string.
// The result may be a cd-prefix path, an action command, or empty if quit.
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
