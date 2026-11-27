package shell

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// fuzzyMatch checks if all chars of query appear in s in order (case-insensitive).
// Returns a score: higher is better. Consecutive matches and match-at-word-start get bonuses.
func fuzzyMatch(s, query string) (bool, int) {
	return scanFuzzyMatch(s, query, nil)
}

// scanFuzzyMatch reports the byte spans of the same subsequence it scores.
func scanFuzzyMatch(s, query string, match func(start, end int)) (bool, int) {
	score := 0
	consecutive, wordStart := false, true
	for i, r := range s {
		if query == "" {
			break
		}
		q, size := utf8.DecodeRuneInString(query)
		if unicode.ToLower(r) != unicode.ToLower(q) {
			consecutive = false
		} else {
			if consecutive {
				score += 5
			}
			if wordStart {
				score += 10
			}
			score++
			consecutive = true
			if match != nil {
				_, width := utf8.DecodeRuneInString(s[i:])
				match(i, i+width)
			}
			query = query[size:]
		}
		wordStart = r == ' ' || r == '-' || r == '_' || r == '/'
	}
	if query != "" {
		return false, 0
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
	focusCmd tea.Cmd
	width    int
	height   int
	done     bool
	result   fuzzySelection
}

type fuzzySelection struct {
	line    string
	execute bool
}

type filteredEntry struct {
	entry string
	score int
}

func (m fuzzyModel) Init() tea.Cmd {
	// The query is focused when the model is constructed (see runFuzzySearch).
	// Focusing here would only mutate this value-receiver copy, leaving the
	// real model's input blurred and unable to accept keystrokes.
	return m.focusCmd
}

func (m fuzzyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.Keystroke() {
		case "esc", "ctrl+g":
			m.done = true
			return m, tea.Quit

		case "enter", "tab":
			m.done = true
			if len(m.filtered) > 0 && m.selected >= 0 && m.selected < len(m.filtered) {
				m.result = fuzzySelection{line: m.filtered[m.selected].entry, execute: msg.Keystroke() == "enter"}
			}
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil

		case "home":
			m.query.SetValue("")
			m.query.SetCursor(0)
			return m.filter()

		case "end":
			m.query.SetCursor(len(m.query.Value()))
			return m, nil

		case "backspace":
			// Let textinput handle it
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			if cmd != nil {
				m, fcmd := m.filter()
				return m, tea.Batch(cmd, fcmd)
			}
			return m.filter()

		default:
			// Let the textinput handle it
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			if cmd != nil {
				m, fcmd := m.filter()
				return m, tea.Batch(cmd, fcmd)
			}
			return m.filter()
		}

	case tea.InterruptMsg:
		m.done = true
		return m, tea.Quit
	}

	// Pass through to textinput for unknown messages
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	if cmd != nil {
		m, fcmd := m.filter()
		return m, tea.Batch(cmd, fcmd)
	}
	return m.filter()
}

func (m fuzzyModel) filter() (fuzzyModel, tea.Cmd) {
	m.filtered = filterEntries(m.entries, m.query.Value())
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 && len(m.filtered) > 0 {
		m.selected = 0
	}

	return m, nil
}

// filterEntries filters and scores entries against a query, returning the
// best matches sorted by descending score (capped at maxResults).
func filterEntries(entries []string, q string) []filteredEntry {
	filtered := make([]filteredEntry, 0, maxResults)
	for _, entry := range entries {
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

	return filtered
}

func (m fuzzyModel) View() tea.View {
	if m.done {
		return tea.View{}
	}

	var b strings.Builder

	// Calculate available lines for results (reserve 2 for input + separator)
	available := max(m.height-3, 0)

	// Show results with scroll around selected entry
	start := max(m.selected-available/2, 0)
	end := start + available
	if end > len(m.filtered) {
		end = len(m.filtered)
		start = max(end-available, 0)
	}

	for i := start; i < end; i++ {
		base, marker := dimStyle, "  "
		if i == m.selected {
			base, marker = selectedStyle, "> "
		}
		row := base.Render(marker) + highlightMatch(m.filtered[i].entry, m.query.Value(), base)
		tail := "…"
		if m.width < 3 {
			tail = ""
		}
		b.WriteString(ansi.Truncate(row, max(m.width, 0), tail))
		b.WriteByte('\n')
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

// Style each span directly so match resets cannot cancel the row background.
func highlightMatch(entry, query string, base lipgloss.Style) string {
	var b strings.Builder
	last := 0
	highlighted := matchStyle(base)
	scanFuzzyMatch(entry, query, func(start, end int) {
		b.WriteString(base.Render(entry[last:start]))
		b.WriteString(highlighted.Render(entry[start:end]))
		last = end
	})
	b.WriteString(base.Render(entry[last:]))
	return b.String()
}

func matchStyle(base lipgloss.Style) lipgloss.Style {
	return base.Foreground(lipgloss.Color("212")).Bold(true).Underline(true)
}

var (
	dimStyle      = lipgloss.NewStyle()
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	inputStyle    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
)

// newFuzzySearchModel builds the fuzzy search model for the given history
// entries. The query input is focused here, when the model is constructed,
// rather than in Init: Init has a value receiver, so focusing there would only
// mutate a discarded copy and leave the running model's input blurred and
// unable to accept keystrokes. The focus command it returns is stored so Init
// can still start the cursor blink.
func newFuzzySearchModel(entries []string) fuzzyModel {
	ti := textinput.New()
	ti.SetWidth(80)
	ti.Prompt = ""
	ti.Placeholder = "type to filter..."
	focusCmd := ti.Focus()

	m := fuzzyModel{
		entries:  entries,
		filtered: make([]filteredEntry, 0),
		selected: 0,
		query:    ti,
		focusCmd: focusCmd,
		width:    80,
		height:   20,
	}

	// Do initial filter
	m.filter()
	return m
}

// runFuzzySearch launches the bubbletea TUI fuzzy search with the given history entries.
// Returns the selected entry and whether to execute it immediately.
func (s *Shell) runFuzzySearch() fuzzySelection {
	entries := s.readHistory()
	if len(entries) == 0 {
		fmt.Fprint(s.Stderr, "\n(fuzzy-search): no history entries\n")
		return fuzzySelection{}
	}

	m := newFuzzySearchModel(entries)

	p := tea.NewProgram(m, s.restoreStdin(), tea.WithOutput(s.Stdout))
	result, err := p.Run()
	if err != nil {
		// Ignore ErrInterrupted (user pressed Ctrl+C)
		return fuzzySelection{}
	}

	fm := result.(fuzzyModel)
	return fm.result
}

// restoreStdin returns a ProgramOption that sets the input to the original stdin
// (after the interruptible pipe was closed).
func (s *Shell) restoreStdin() tea.ProgramOption {
	return tea.WithInput(s.Stdin)
}
