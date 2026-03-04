package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sergidb/wt/internal/config"
)

type configScreen int

const (
	cfgScreenList configScreen = iota
	cfgScreenEdit
	cfgScreenDelete
)

var availableColors = []string{"green", "cyan", "yellow", "magenta", "blue", "red", "white"}

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
				m.statusMsg = errorStyle.Render("Error: " + err.Error())
				return m, nil
			}
			m.statusMsg = successStyle.Render("Saved!")
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
				m.statusMsg = errorStyle.Render("Error: " + err.Error())
			} else {
				m.statusMsg = successStyle.Render("Deleted.")
			}
		}
		m.screen = cfgScreenList
	}
	return m, nil
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
	}
	return ""
}

func (m configModel) viewList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Services (.wt.yaml)"))
	b.WriteString("\n")

	if len(m.serviceNames) == 0 {
		b.WriteString(dimStyle.Render("  No services configured yet."))
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

			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("> ")
			}

			displayName := name + strings.Repeat(" ", maxLen-len(name))
			if i == m.cursor {
				displayName = selectedStyle.Render(displayName)
			}

			cmd := dimStyle.Render("  " + svc.Cmd)

			b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, displayName, cmd))
		}
	}

	if m.statusMsg != "" {
		b.WriteString("\n  " + m.statusMsg + "\n")
	}

	b.WriteString(helpStyle.Render("a add • e/enter edit • d delete • q back"))

	return b.String()
}

func (m configModel) viewEdit() string {
	var b strings.Builder

	title := "Edit Service"
	if m.isNew {
		title = "Add Service"
	}
	b.WriteString(titleStyle.Render("  " + title))
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
		// Skip name field when editing
		if !m.isNew && f.idx == fieldName {
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render(f.label+":"), m.inputs[f.idx].Value()))
			continue
		}

		active := ""
		if f.idx == m.inputIdx {
			active = " "
		}

		hint := ""
		if f.hint != "" {
			hint = dimStyle.Render(" (" + f.hint + ")")
		}

		b.WriteString(fmt.Sprintf("  %s%s%s\n", active, labelStyle.Render(f.label+":"), hint))
		b.WriteString(fmt.Sprintf("  %s\n", m.inputs[f.idx].View()))
	}

	if m.statusMsg != "" {
		b.WriteString("\n  " + m.statusMsg + "\n")
	}

	b.WriteString(helpStyle.Render("tab/enter next • shift+tab prev • esc cancel"))

	return b.String()
}

func (m configModel) viewDelete() string {
	var b strings.Builder

	name := m.serviceNames[m.deleteIdx]
	b.WriteString(titleStyle.Render(fmt.Sprintf("  Delete service '%s'?", name)))
	b.WriteString("\n")

	options := []string{"Yes", "No"}
	for i, opt := range options {
		cursor := "  "
		if i == m.deleteCursor {
			cursor = cursorStyle.Render("> ")
		}

		label := opt
		if i == m.deleteCursor {
			label = selectedStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, label))
	}

	b.WriteString(helpStyle.Render("↑/↓ select • enter confirm • esc cancel"))

	return b.String()
}

// RunConfig launches the config TUI.
func RunConfig(repoRoot string) error {
	m := newConfigModel(repoRoot)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	_, err := p.Run()
	return err
}
