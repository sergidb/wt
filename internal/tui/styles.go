package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	sourceGitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35"))

	sourceClaudeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208"))

	mainBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)
