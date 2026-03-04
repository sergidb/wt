package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergidb/wt/internal/config"
)

type configScreen int

const (
	cfgScreenList configScreen = iota
	cfgScreenEdit
	cfgScreenDelete
	cfgScreenWorktreesDir
)

var availableColors = []string{"green", "cyan", "yellow", "magenta", "blue", "red", "white"}

// Color codes for preview blocks
var colorPreview = map[string]lipgloss.Color{
	"green":   lipgloss.Color("78"),
	"cyan":    lipgloss.Color("39"),
	"yellow":  lipgloss.Color("214"),
	"magenta": lipgloss.Color("213"),
	"blue":    lipgloss.Color("69"),
	"red":     lipgloss.Color("196"),
	"white":   lipgloss.Color("255"),
}

// Form field indices
const (
	fieldName = iota
	fieldCmd
	fieldDir
	fieldColor
	fieldCount
)

type configModel struct {
	cfg          *config.Config
	repoRoot     string
	serviceNames []string
	cursor       int

	screen configScreen

	// Edit/Add form
	inputs   []textinput.Model
	inputIdx int
	isNew    bool

	// Delete confirmation
	deleteIdx    int
	deleteCursor int

	// Color picker
	colorIdx int

	// Worktrees dir edit
	wtDirInput textinput.Model

	// Status message
	statusMsg string

	quitting bool
}

func newConfigModel(repoRoot string) configModel {
	cfg := config.LoadOrEmpty(repoRoot)

	m := configModel{
		cfg:      cfg,
		repoRoot: repoRoot,
	}
	m.refreshNames()
	m.initInputs()
	m.initWtDirInput()
	return m
}

func (m *configModel) refreshNames() {
	m.serviceNames = make([]string, 0, len(m.cfg.Services))
	for name := range m.cfg.Services {
		m.serviceNames = append(m.serviceNames, name)
	}
	sort.Strings(m.serviceNames)
}

func (m *configModel) initInputs() {
	m.inputs = make([]textinput.Model, fieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "service name"
	nameInput.CharLimit = 30
	m.inputs[fieldName] = nameInput

	cmdInput := textinput.New()
	cmdInput.Placeholder = "e.g. npm run dev"
	cmdInput.CharLimit = 200
	m.inputs[fieldCmd] = cmdInput

	dirInput := textinput.New()
	dirInput.Placeholder = "."
	dirInput.CharLimit = 100
	m.inputs[fieldDir] = dirInput

	colorInput := textinput.New()
	colorInput.Placeholder = "auto"
	colorInput.CharLimit = 20
	m.inputs[fieldColor] = colorInput
}

func (m *configModel) initWtDirInput() {
	wtDir := textinput.New()
	wtDir.Placeholder = ".worktrees"
	wtDir.CharLimit = 200
	wtDir.SetValue(m.cfg.WorktreesDir)
	m.wtDirInput = wtDir
}

func (m *configModel) focusInput(idx int) {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	m.inputIdx = idx
	m.inputs[idx].Focus()
}

func (m *configModel) startAdd() {
	m.screen = cfgScreenEdit
	m.isNew = true
	m.statusMsg = ""
	m.colorIdx = 0

	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	m.focusInput(fieldName)
}

func (m *configModel) startEdit() {
	if len(m.serviceNames) == 0 {
		return
	}
	name := m.serviceNames[m.cursor]
	svc := m.cfg.Services[name]

	m.screen = cfgScreenEdit
	m.isNew = false
	m.statusMsg = ""

	m.inputs[fieldName].SetValue(name)
	m.inputs[fieldCmd].SetValue(svc.Cmd)
	m.inputs[fieldDir].SetValue(svc.Dir)
	m.inputs[fieldColor].SetValue(svc.Color)

	// Find color index
	m.colorIdx = 0
	for i, c := range availableColors {
		if c == svc.Color {
			m.colorIdx = i
			break
		}
	}

	m.focusInput(fieldCmd) // skip name field when editing
}

func (m *configModel) saveService() error {
	name := strings.TrimSpace(m.inputs[fieldName].Value())
	cmd := strings.TrimSpace(m.inputs[fieldCmd].Value())
	dir := strings.TrimSpace(m.inputs[fieldDir].Value())
	color := strings.TrimSpace(m.inputs[fieldColor].Value())

	if name == "" {
		return fmt.Errorf("name is required")
	}
	if cmd == "" {
		return fmt.Errorf("command is required")
	}
	if dir == "" {
		dir = "."
	}

	if m.cfg.Services == nil {
		m.cfg.Services = make(map[string]config.Service)
	}

	m.cfg.Services[name] = config.Service{
		Cmd:   cmd,
		Dir:   dir,
		Color: color,
	}

	if err := config.Save(m.repoRoot, m.cfg); err != nil {
		return err
	}

	m.refreshNames()
	return nil
}

func (m *configModel) deleteService() error {
	if len(m.serviceNames) == 0 {
		return nil
	}
	name := m.serviceNames[m.deleteIdx]
	delete(m.cfg.Services, name)

	if err := config.Save(m.repoRoot, m.cfg); err != nil {
		return err
	}

	m.refreshNames()
	if m.cursor >= len(m.serviceNames) && m.cursor > 0 {
		m.cursor--
	}
	return nil
}

func (m configModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case cfgScreenList:
			return m.updateList(msg)
		case cfgScreenEdit:
			return m.updateEdit(msg)
		case cfgScreenDelete:
			return m.updateDelete(msg)
		case cfgScreenWorktreesDir:
			return m.updateWorktreesDir(msg)
		}
	}

	// Update active text input
	if m.screen == cfgScreenEdit {
		var cmd tea.Cmd
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m configModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.serviceNames)-1 {
			m.cursor++
		}
	case "a":
		m.startAdd()
		return m, textinput.Blink
	case "e", "enter":
		m.startEdit()
		return m, textinput.Blink
	case "d":
		if len(m.serviceNames) > 0 {
			m.screen = cfgScreenDelete
			m.deleteIdx = m.cursor
			m.deleteCursor = 1 // default to "No"
		}
	case "w":
		m.screen = cfgScreenWorktreesDir
		m.statusMsg = ""
		m.wtDirInput.SetValue(m.cfg.WorktreesDir)
		m.wtDirInput.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m configModel) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = cfgScreenList
		m.statusMsg = ""
		return m, nil
	case "tab", "enter":
		// Move to next field or save
		nextIdx := m.inputIdx + 1

		// Skip name field when editing
		if !m.isNew && nextIdx == fieldName {
			nextIdx++
		}

		if nextIdx >= fieldCount {
			// Save
			if err := m.saveService(); err != nil {
				m.statusMsg = errorStyle.Render("  Error: " + err.Error())
				return m, nil
			}
			m.statusMsg = successStyle.Render("  Saved!")
			m.screen = cfgScreenList
			return m, nil
		}

		m.focusInput(nextIdx)
		return m, textinput.Blink
	case "shift+tab":
		// Move to previous field
		prevIdx := m.inputIdx - 1
		if !m.isNew && prevIdx == fieldName {
			prevIdx--
		}
		if prevIdx < 0 || (!m.isNew && prevIdx < fieldCmd) {
			return m, nil
		}
		m.focusInput(prevIdx)
		return m, textinput.Blink
	}

	// Handle color field specially - cycle through colors
	if m.inputIdx == fieldColor {
		switch msg.String() {
		case "left", "right":
			if msg.String() == "right" {
				m.colorIdx = (m.colorIdx + 1) % len(availableColors)
			} else {
				m.colorIdx = (m.colorIdx - 1 + len(availableColors)) % len(availableColors)
			}
			m.inputs[fieldColor].SetValue(availableColors[m.colorIdx])
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
	return m, cmd
}

func (m configModel) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = cfgScreenList
	case "up", "k":
		if m.deleteCursor > 0 {
			m.deleteCursor--
		}
	case "down", "j":
		if m.deleteCursor < 1 {
			m.deleteCursor++
		}
	case "enter":
		if m.deleteCursor == 0 {
			// Yes - delete
			if err := m.deleteService(); err != nil {
				m.statusMsg = errorStyle.Render("  Error: " + err.Error())
			} else {
				m.statusMsg = successStyle.Render("  Deleted.")
			}
		}
		m.screen = cfgScreenList
	}
	return m, nil
}

func (m configModel) updateWorktreesDir(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = cfgScreenList
		m.wtDirInput.Blur()
		return m, nil
	case "enter":
		m.cfg.WorktreesDir = strings.TrimSpace(m.wtDirInput.Value())
		if err := config.Save(m.repoRoot, m.cfg); err != nil {
			m.statusMsg = errorStyle.Render("  Error: " + err.Error())
		} else {
			m.statusMsg = successStyle.Render("  Saved!")
		}
		m.wtDirInput.Blur()
		m.screen = cfgScreenList
		return m, nil
	}

	var cmd tea.Cmd
	m.wtDirInput, cmd = m.wtDirInput.Update(msg)
	return m, cmd
}

func (m configModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case cfgScreenList:
		return m.viewList()
	case cfgScreenEdit:
		return m.viewEdit()
	case cfgScreenDelete:
		return m.viewDelete()
	case cfgScreenWorktreesDir:
		return m.viewWorktreesDir()
	}
	return ""
}

func (m configModel) viewList() string {
	var b strings.Builder

	// Header
	b.WriteString(renderHeader("Config"))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	// Worktrees directory setting
	wtDirDisplay := m.cfg.WorktreesDir
	if wtDirDisplay == "" {
		wtDirDisplay = dimStyle.Render(".worktrees (default)")
	} else {
		wtDirDisplay = secondaryStyle.Render(wtDirDisplay)
	}
	b.WriteString(fmt.Sprintf("   %s  %s\n", labelStyle.Render("worktrees_dir:"), wtDirDisplay))
	b.WriteString("\n")

	// Services section
	b.WriteString(labelStyle.Render("   Services"))
	b.WriteString("\n\n")

	if len(m.serviceNames) == 0 {
		b.WriteString(dimStyle.Render("   No services configured yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		// Find max name length for alignment
		maxLen := 0
		for _, name := range m.serviceNames {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}

		for i, name := range m.serviceNames {
			svc := m.cfg.Services[name]
			isSelected := i == m.cursor

			cursor := "   "
			if isSelected {
				cursor = cursorStyle.Render(" ▸ ")
			}

			displayName := name + strings.Repeat(" ", maxLen-len(name))
			if isSelected {
				displayName = selectedStyle.Render(displayName)
			} else {
				displayName = secondaryStyle.Render(displayName)
			}

			// Color indicator
			colorDot := ""
			if svc.Color != "" {
				if c, ok := colorPreview[svc.Color]; ok {
					colorDot = lipgloss.NewStyle().Foreground(c).Render("●") + " "
				}
			}

			cmd := dimStyle.Render(svc.Cmd)

			b.WriteString(fmt.Sprintf("%s%s%s  %s\n", cursor, colorDot, displayName, cmd))

			// Show dir if not default
			if isSelected && svc.Dir != "" && svc.Dir != "." {
				b.WriteString(dimStyle.Render(fmt.Sprintf("     dir: %s", svc.Dir)) + "\n")
			}
		}
	}

	// Status
	if m.statusMsg != "" {
		b.WriteString("\n" + m.statusMsg + "\n")
	}

	// Status bar
	count := fmt.Sprintf("%d services", len(m.serviceNames))
	b.WriteString(renderStatusBar(count, ""))

	// Separator + Help
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("a add  e/enter edit  d delete  w worktrees dir  q back"))

	return b.String()
}

func (m configModel) viewEdit() string {
	var b strings.Builder

	// Header
	title := "Edit Service"
	if m.isNew {
		title = "Add Service"
	}
	b.WriteString(renderHeader("Config", title))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	fields := []struct {
		label string
		idx   int
		hint  string
	}{
		{"Name", fieldName, ""},
		{"Command", fieldCmd, "required"},
		{"Directory", fieldDir, "relative to worktree root"},
		{"Color", fieldColor, "←/→ to cycle"},
	}

	for _, f := range fields {
		isActive := f.idx == m.inputIdx

		// Skip name field when editing (show as read-only header)
		if !m.isNew && f.idx == fieldName {
			b.WriteString(fmt.Sprintf("   %s %s\n\n", labelStyle.Render(f.label+":"), selectedStyle.Render(m.inputs[f.idx].Value())))
			continue
		}

		// Active indicator
		indicator := fieldInactiveStyle.String()
		if isActive {
			indicator = fieldActiveStyle.String()
		}

		// Label + hint
		label := labelStyle.Render(f.label + ":")
		hint := ""
		if f.hint != "" {
			hint = dimStyle.Render(" (" + f.hint + ")")
		}

		b.WriteString(fmt.Sprintf(" %s%s%s\n", indicator, label, hint))

		// Input field
		b.WriteString(fmt.Sprintf("     %s\n", m.inputs[f.idx].View()))

		// Color preview
		if f.idx == fieldColor && m.inputs[fieldColor].Value() != "" {
			colorName := m.inputs[fieldColor].Value()
			if c, ok := colorPreview[colorName]; ok {
				preview := lipgloss.NewStyle().Foreground(c).Render("████")
				b.WriteString(fmt.Sprintf("     %s %s\n", preview, dimStyle.Render(colorName)))
			}
		}

		b.WriteString("\n")
	}

	// Status
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}

	// Separator + Help
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("tab/enter next  shift+tab prev  esc cancel"))

	return b.String()
}

func (m configModel) viewDelete() string {
	var b strings.Builder

	name := m.serviceNames[m.deleteIdx]

	// Header
	b.WriteString(renderHeader("Config", "Delete"))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	b.WriteString(warningStyle.Render(fmt.Sprintf("   Delete service '%s'?", name)))
	b.WriteString("\n\n")

	options := []string{"Yes, delete", "No, cancel"}
	for i, opt := range options {
		isSelected := i == m.deleteCursor

		cursor := "   "
		if isSelected {
			cursor = cursorStyle.Render(" ▸ ")
		}

		label := opt
		if isSelected {
			if i == 0 {
				label = errorStyle.Render(opt)
			} else {
				label = selectedStyle.Render(opt)
			}
		} else {
			label = secondaryStyle.Render(opt)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, label))
	}

	// Separator + Help
	b.WriteString("\n")
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("↑↓ select  enter confirm  esc cancel"))

	return b.String()
}

func (m configModel) viewWorktreesDir() string {
	var b strings.Builder

	b.WriteString(renderHeader("Config", "Worktrees Directory"))
	b.WriteString(renderSeparator(50))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("   %s\n", labelStyle.Render("Path:")))
	b.WriteString(fmt.Sprintf("     %s\n\n", m.wtDirInput.View()))

	b.WriteString(dimStyle.Render("   Relative to repo root, or absolute path."))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("   Leave empty for default (.worktrees)."))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(renderSeparator(50))
	b.WriteString(renderHelp("enter save  esc cancel"))

	return b.String()
}

// RunConfig launches the config TUI.
func RunConfig(repoRoot string) error {
	m := newConfigModel(repoRoot)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	_, err := p.Run()
	return err
}
