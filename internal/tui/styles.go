package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Digit4l brand palette — 256-color codes for max compatibility
var (
	colorPrimary   = lipgloss.Color("25")  // Deep blue (~#0F3BAB)
	colorSecondary = lipgloss.Color("68")  // Medium blue (~#5C83CC)
	colorWhite     = lipgloss.Color("255") // White — main text
	colorSuccess   = lipgloss.Color("114") // Green
	colorWarning   = lipgloss.Color("214") // Amber
	colorError     = lipgloss.Color("203") // Red
	colorLightGray = lipgloss.Color("252") // Light gray
	colorDim       = lipgloss.Color("245") // Medium gray
	colorMuted     = lipgloss.Color("240") // Dark gray
	colorManaged = lipgloss.Color("173") // Managed worktree accent
)

// Brand — D4 signature
var (
	d4Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	breadcrumbSep = lipgloss.NewStyle().
			Foreground(colorMuted).
			SetString(" › ")

	breadcrumbActive = lipgloss.NewStyle().
				Foreground(colorWhite).
				Bold(true)

	breadcrumbDim = lipgloss.NewStyle().
			Foreground(colorDim)
)

// Layout
var (
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// List items
var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	textStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	secondaryStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)
)

// Badges
var (
	badgeMainStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	badgeGitStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	badgeManagedStyle = lipgloss.NewStyle().
				Foreground(colorManaged)
)

// Actions
var (
	actionNameStyle = lipgloss.NewStyle().
			Width(10)

	actionDescStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// Forms
var (
	labelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	fieldActiveStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				SetString("▸ ")

	fieldInactiveStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				SetString("  ")

	inputStyle = lipgloss.NewStyle().
			Foreground(colorWhite)
)

// Feedback
var (
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)
)

// Helper functions

func renderHeader(parts ...string) string {
	var b strings.Builder
	b.WriteString("\n ")
	b.WriteString(d4Style.Render("D4"))
	b.WriteString(" ")
	b.WriteString(brandStyle.Render("wt"))
	for i, part := range parts {
		b.WriteString(breadcrumbSep.String())
		if i == len(parts)-1 {
			b.WriteString(breadcrumbActive.Render(part))
		} else {
			b.WriteString(breadcrumbDim.Render(part))
		}
	}
	b.WriteString("\n")
	return b.String()
}

func renderSeparator(width int) string {
	if width < 1 {
		width = 40
	}
	return separatorStyle.Render(" " + strings.Repeat("─", width)) + "\n"
}

func renderHelp(text string) string {
	return "\n" + helpStyle.Render(" "+text) + "\n"
}

func renderStatusBar(left string, right string) string {
	if left == "" && right == "" {
		return ""
	}
	result := " "
	if left != "" {
		result += statusBarStyle.Render(left)
	}
	if right != "" {
		if left != "" {
			result += "  "
		}
		result += right
	}
	return result + "\n"
}
