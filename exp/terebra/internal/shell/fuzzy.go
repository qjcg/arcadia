package shell

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// fuzzyMatch checks if all chars of query appear in s in order (case-insensitive).
// Returns a score: higher is better. Consecutive matches and match-at-word-start get bonuses.
func fuzzyMatch(s, query string) (bool, int) {
	s = strings.ToLower(s)
	ql := strings.ToLower(query)
	if ql == "" {
		return true, 0
	}

	si := 0
	score := 0
	prevMatched := false
	for _, qc := range ql {
		found := false
		for si < len(s) {
			sc := rune(s[si])
			si++
			if sc == qc {
				found = true
				// Bonus for consecutive matches
				if prevMatched {
					score += 5
				}
				// Bonus for match at word start
				if si == 1 || (si > 1 && (s[si-2] == ' ' || s[si-2] == '-' || s[si-2] == '_' || s[si-2] == '/')) {
					score += 10
				}
				prevMatched = true
				score++
				break
			}
			prevMatched = false
		}
		if !found {
			return false, 0
		}
	}
	return true, score
}

// maxResults is the maximum number of results to display.
const maxResults = 50

type fuzzyModel struct {
	entries  []string
	filtered []filteredEntry
	selected int
	query    textinput.Model
	width    int
	height   int
	done     bool
	result   string
}

type filteredEntry struct {
	entry string
	score int
}

func (m fuzzyModel) Init() tea.Cmd {
	return m.query.Focus()
}

func (m fuzzyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEscape:
			m.done = true
			m.result = ""
			return m, tea.Quit

		case tea.KeyEnter:
			m.done = true
			if len(m.filtered) > 0 && m.selected >= 0 && m.selected < len(m.filtered) {
				m.result = m.filtered[m.selected].entry
			}
			return m, tea.Quit

		case tea.KeyUp:
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case tea.KeyDown:
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil

		case tea.KeyHome:
			m.query.SetValue("")
			m.query.SetCursor(0)
			return m, m.filter()

		case tea.KeyEnd:
			m.query.SetCursor(len(m.query.Value()))
			return m, nil

		case tea.KeyBackspace:
			// Let textinput handle it
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			if cmd != nil {
				return m, tea.Batch(cmd, m.filter())
			}
			return m, m.filter()

		default:
			// Let the textinput handle it
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			if cmd != nil {
				return m, tea.Batch(cmd, m.filter())
			}
			return m, m.filter()
		}

	case tea.InterruptMsg:
		m.done = true
		m.result = ""
		return m, tea.Quit
	}

	// Pass through to textinput for unknown messages
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	if cmd != nil {
		return m, tea.Batch(cmd, m.filter())
	}
	return m, m.filter()
}

func (m fuzzyModel) filter() tea.Cmd {
	q := m.query.Value()
	filtered := make([]filteredEntry, 0, maxResults)
	for _, entry := range m.entries {
		if ok, score := fuzzyMatch(entry, q); ok {
			filtered = append(filtered, filteredEntry{entry, score})
		}
	}

	// Sort by score descending (simple insertion sort)
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].score > filtered[i].score {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if len(filtered) > maxResults {
		filtered = filtered[:maxResults]
	}

	m.filtered = filtered
	if m.selected >= len(filtered) {
		m.selected = len(filtered) - 1
	}
	if m.selected < 0 && len(filtered) > 0 {
		m.selected = 0
	}

	return nil
}

func (m fuzzyModel) View() tea.View {
	if m.done {
		return tea.View{}
	}

	var b strings.Builder

	// Calculate available lines for results (reserve 2 for input + separator)
	available := m.height - 3
	if available < 0 {
		available = 0
	}

	// Show results with scroll around selected entry
	start := m.selected - available/2
	if start < 0 {
		start = 0
	}
	end := start + available
	if end > len(m.filtered) {
		end = len(m.filtered)
		start = end - available
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		entry := m.filtered[i].entry

		// Truncate to width
		if len(entry) > m.width-4 {
			entry = entry[:m.width-4] + "..."
		}

		// Highlight the matching characters in the query
		if i == m.selected {
			disp := highlightMatch(entry, m.query.Value())
			b.WriteString(selectedStyle.Render(disp))
			b.WriteByte('\n')
		} else {
			disp := highlightMatch(entry, m.query.Value())
			b.WriteString(dimStyle.Render(disp))
			b.WriteByte('\n')
		}
	}

	// Pad remaining lines
	for i := end - start; i < available; i++ {
		b.WriteByte('\n')
	}

	// Separator line
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteByte('\n')

	// Input line
	b.WriteString(inputStyle.Render("> "))
	b.WriteString(m.query.View())

	v := tea.NewView(b.String())
	v.AltScreen = m.width > 0 && m.height > 0
	return v
}

// highlightMatch returns a string with matching characters styled.
func highlightMatch(entry, query string) string {
	if query == "" {
		return entry
	}

	ql := strings.ToLower(query)
	el := strings.ToLower(entry)

	var b strings.Builder
	ei := 0
	for _, qc := range ql {
		for ei < len(el) {
			rc := rune(el[ei])
			if rc == qc {
				b.WriteString(highlightStyle.Render(string(entry[ei])))
				ei++
				break
			}
			b.WriteByte(entry[ei])
			ei++
		}
	}
	// Write remaining characters
	for ei < len(entry) {
		b.WriteByte(entry[ei])
		ei++
	}
	return b.String()
}

var (
	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("236"))
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
)

// runFuzzySearch launches the bubbletea TUI fuzzy search with the given history entries.
// Returns the selected entry, or empty string if cancelled.
func (s *Shell) runFuzzySearch() string {
	entries := s.readHistory()
	if len(entries) == 0 {
		fmt.Fprint(s.Stderr, "\n(fuzzy-search): no history entries\n")
		return ""
	}

	ti := textinput.New()
	ti.SetWidth(80)
	ti.Prompt = ""
	ti.Placeholder = "type to filter..."

	m := fuzzyModel{
		entries:  entries,
		filtered: make([]filteredEntry, 0),
		selected: 0,
		query:    ti,
		width:    80,
		height:   20,
	}

	// Do initial filter
	m.filter()

	p := tea.NewProgram(m, s.restoreStdin(), tea.WithOutput(s.Stdout))
	result, err := p.Run()
	if err != nil {
		// Ignore ErrInterrupted (user pressed Ctrl+C)
		return ""
	}

	fm := result.(fuzzyModel)
	return fm.result
}

// restoreStdin returns a ProgramOption that sets the input to the original stdin
// (after the interruptible pipe was closed).
func (s *Shell) restoreStdin() tea.ProgramOption {
	return tea.WithInput(s.Stdin)
}
