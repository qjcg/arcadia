package scaffold

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PromptForVariables asks the user for each variable interactively using
// bubbletea (textinput for freeform, list for choices).
// In quiet mode, returns defaults for everything.
func PromptForVariables(vars []Variable, quiet bool) map[string]string {
	values := make(map[string]string, len(vars))
	if quiet {
		for _, v := range vars {
			values[v.Name] = v.Default
		}
		return values
	}

	steps := buildSteps(vars)
	m := newWizard(steps)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	wiz := result.(*wizard)
	if wiz.cancelled {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)
	}
	return wiz.values
}

// --- wizard model ---

type step struct {
	name     string
	prompt   string
	varType  VariableType
	defaultV string
	choices  []string
	required bool
}

func buildSteps(vars []Variable) []step {
	steps := make([]step, len(vars))
	for i, v := range vars {
		steps[i] = step{
			name:     v.Name,
			prompt:   v.Prompt,
			varType:  v.Type,
			defaultV: v.Default,
			choices:  v.Choices,
			required: v.Required,
		}
	}
	return steps
}

var (
	styleHeader   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	stylePrompt   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	styleRequired = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleStep     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

// choiceItem implements list.Item interface for bubbles v2.
type choiceItem struct{ title string }

func (c choiceItem) Title() string       { return c.title }
func (c choiceItem) Description() string { return "" }
func (c choiceItem) FilterValue() string { return c.title }

// --- bubbletea model ---

type wizard struct {
	steps     []step
	current   int
	values    map[string]string
	input     textinput.Model
	list      list.Model
	inList    bool
	width     int
	height    int
	cancelled bool
}

func newWizard(steps []step) *wizard {
	w := &wizard{
		steps:  steps,
		values: make(map[string]string, len(steps)),
	}
	if len(steps) > 0 {
		w.initCurrent()
	}
	return w
}

func (w *wizard) initCurrent() {
	if w.current >= len(w.steps) {
		return
	}
	s := w.steps[w.current]

	switch s.varType {
	case TypeFreeform:
		w.inList = false
		ti := textinput.New()
		ti.SetWidth(50)
		ti.Placeholder = s.defaultV
		if s.defaultV != "" {
			ti.SetValue(s.defaultV)
		}
		w.input = ti
		w.input.Focus()

	case TypeChoice:
		w.inList = true
		var items []list.Item
		defaultIdx := 0
		for i, c := range s.choices {
			items = append(items, choiceItem{title: c})
			if c == s.defaultV {
				defaultIdx = i
			}
		}
		delegate := list.NewDefaultDelegate()
		l := list.New(items, delegate, 0, 0)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetShowPagination(false)
		l.SetShowFilter(false)
		l.Select(defaultIdx)
		w.list = l
	}
}

func (w *wizard) Init() tea.Cmd {
	if len(w.steps) == 0 {
		return tea.Quit
	}
	if w.inList {
		return nil
	}
	return textinput.Blink
}

func (w *wizard) nextStep() tea.Cmd {
	w.current++
	if w.current >= len(w.steps) {
		return tea.Quit
	}
	w.initCurrent()
	if w.inList {
		return nil
	}
	return textinput.Blink
}

func (w *wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if w.current >= len(w.steps) {
		return w, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Code == 'c' && msg.Mod == tea.ModCtrl {
			w.cancelled = true
			return w, tea.Quit
		}
		if msg.Code == tea.KeyEnter {
			s := w.steps[w.current]
			if w.inList {
				selected, ok := w.list.SelectedItem().(choiceItem)
				if ok {
					w.values[s.name] = selected.title
				} else if s.defaultV != "" {
					w.values[s.name] = s.defaultV
				}
				return w, w.nextStep()
			}
			val := w.input.Value()
			if val == "" && s.defaultV != "" {
				val = s.defaultV
			}
			if val == "" && s.required {
				return w, nil
			}
			w.values[s.name] = val
			return w, w.nextStep()
		}

	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		if w.inList {
			availH := msg.Height - 4
			if availH < 3 {
				availH = 3
			}
			w.list.SetSize(msg.Width-4, availH)
		}
	}

	if w.inList {
		var cmd tea.Cmd
		w.list, cmd = w.list.Update(msg)
		return w, cmd
	}
	var cmd tea.Cmd
	w.input, cmd = w.input.Update(msg)
	return w, cmd
}

func (w *wizard) View() tea.View {
	if w.current >= len(w.steps) {
		return tea.NewView("")
	}
	s := w.steps[w.current]

	var stepInfo string
	if len(w.steps) > 1 {
		stepInfo = styleStep.Render(fmt.Sprintf("(%d/%d)", w.current+1, len(w.steps)))
	}

	req := ""
	if s.required {
		req = styleRequired.Render(" (required)")
	}

	header := styleHeader.Render("pavona") + " " + stepInfo
	prompt := stylePrompt.Render(s.prompt) + req

	var body string
	if w.inList {
		body = "\n" + w.list.View()
	} else {
		body = "\n" + w.input.View()
	}

	if w.width > 0 {
		return tea.NewView(
			lipgloss.NewStyle().Width(w.width).Padding(1, 2).Render(
				header + "\n\n" + prompt + body,
			),
		)
	}
	return tea.NewView(header + "\n\n" + prompt + body + "\n")
}
