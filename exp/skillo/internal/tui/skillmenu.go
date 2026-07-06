// Package tui provides interactive TUI components for skillo.
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SkillInfo describes a skill shown in the menu.
type SkillInfo struct {
	Name        string
	Description string
	Installed   bool // whether the skill is already installed
}

// RunSkillMenu shows an interactive checkbox menu and returns the selected skill
// names. Returns nil if the user cancelled.
func RunSkillMenu(title string, skills []SkillInfo) ([]string, error) {
	checked := make(map[string]bool, len(skills))
	for _, s := range skills {
		checked[s.Name] = s.Installed
	}
	p := tea.NewProgram(model{
		title:   title,
		skills:  skills,
		checked: checked,
	})
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	result := m.(model)
	if result.cancelled {
		return nil, nil
	}
	var selected []string
	for _, s := range skills {
		if result.checked[s.Name] {
			selected = append(selected, s.Name)
		}
	}
	return selected, nil
}

type model struct {
	title     string
	skills    []SkillInfo
	checked   map[string]bool
	cursor    int
	cancelled bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'q', tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.skills)-1 {
				m.cursor++
			}
		case ' ':
			name := m.skills[m.cursor].Name
			m.checked[name] = !m.checked[name]
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.title))
	b.WriteByte('\n')

	if len(m.skills) == 0 {
		b.WriteString("\n  No skills available.\n\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter or q to quit."))
		return tea.NewView(b.String())
	}

	for i, s := range m.skills {
		b.WriteByte('\n')
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render("▸ ")
		}
		mark := " "
		if m.checked[s.Name] {
			mark = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Render("x")
		}
		fmt.Fprintf(&b, "%s[%s] %s", cursor, mark, s.Name)
		if s.Description != "" {
			desc := lipgloss.NewStyle().Faint(true).Render("  " + s.Description)
			b.WriteString(desc)
		}
	}
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓ navigate • Space toggle • Enter confirm • q quit"))
	return tea.NewView(b.String())
}
