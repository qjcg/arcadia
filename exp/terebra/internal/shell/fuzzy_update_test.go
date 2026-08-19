package shell

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// newFuzzyModel builds a fuzzyModel with the given entries and an initial query.
func newFuzzyModel(entries []string, query string) fuzzyModel {
	ti := textinput.New()
	ti.SetValue(query)
	ti.Focus()
	m := fuzzyModel{
		entries:  entries,
		filtered: make([]filteredEntry, 0),
		selected: 0,
		query:    ti,
		width:    80,
		height:   20,
	}
	m, _ = m.filter()
	return m
}

func keyPress(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

func TestFuzzyUpdateEscape(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b"}, "")
	nm, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done after escape")
	}
	if fm.result != "" {
		t.Fatalf("expected empty result, got %q", fm.result)
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestFuzzyUpdateEnterSelects(t *testing.T) {
	m := newFuzzyModel([]string{"alpha", "beta"}, "")
	nm, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done after enter")
	}
	if fm.result != "alpha" {
		t.Fatalf("expected 'alpha', got %q", fm.result)
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestFuzzyUpdateEnterNoResults(t *testing.T) {
	m := newFuzzyModel([]string{"alpha"}, "zzz")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done")
	}
	if fm.result != "" {
		t.Fatalf("expected empty result with no matches, got %q", fm.result)
	}
}

func TestFuzzyUpdateArrowDown(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b", "c"}, "")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowDownAtEnd(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b"}, "")
	m.selected = 1
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected to stay 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowUp(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b", "c"}, "")
	m.selected = 2
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowUpAtTop(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b"}, "")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	fm := nm.(fuzzyModel)
	if fm.selected != 0 {
		t.Fatalf("expected selected to stay 0, got %d", fm.selected)
	}
}

func TestFuzzyUpdateHomeClearsQuery(t *testing.T) {
	m := newFuzzyModel([]string{"a", "b"}, "a")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	fm := nm.(fuzzyModel)
	if fm.query.Value() != "" {
		t.Fatalf("expected empty query after home, got %q", fm.query.Value())
	}
}

func TestFuzzyUpdateTypingFilters(t *testing.T) {
	m := newFuzzyModel([]string{"alpha", "beta", "gamma"}, "")
	nm, _ := m.Update(keyPress('a'))
	fm := nm.(fuzzyModel)
	if fm.query.Value() != "a" {
		t.Fatalf("expected query 'a', got %q", fm.query.Value())
	}
	// alpha, beta, and gamma all contain 'a'
	if len(fm.filtered) != 3 {
		t.Fatalf("expected 3 filtered entries, got %d", len(fm.filtered))
	}
}

func TestFuzzyUpdateBackspace(t *testing.T) {
	m := newFuzzyModel([]string{"alpha", "beta"}, "al")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	fm := nm.(fuzzyModel)
	if fm.query.Value() != "a" {
		t.Fatalf("expected query 'a' after backspace, got %q", fm.query.Value())
	}
}

func TestFuzzyUpdateWindowSize(t *testing.T) {
	m := newFuzzyModel([]string{"a"}, "")
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fm := nm.(fuzzyModel)
	if fm.width != 100 || fm.height != 30 {
		t.Fatalf("expected (100, 30), got (%d, %d)", fm.width, fm.height)
	}
}

func TestFuzzyUpdateInterrupt(t *testing.T) {
	m := newFuzzyModel([]string{"a"}, "")
	nm, cmd := m.Update(tea.InterruptMsg{})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done after interrupt")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestFuzzyViewDone(t *testing.T) {
	m := newFuzzyModel([]string{"a"}, "")
	m.done = true
	if got := m.View().Content; got != "" {
		t.Fatalf("expected empty view when done, got %q", got)
	}
}

func TestFuzzyViewRendersEntries(t *testing.T) {
	m := newFuzzyModel([]string{"alpha", "beta"}, "")
	view := m.View().Content
	if !strings.Contains(view, "alpha") {
		t.Fatalf("expected 'alpha' in view, got %q", view)
	}
	if !strings.Contains(view, "beta") {
		t.Fatalf("expected 'beta' in view, got %q", view)
	}
}

func TestFuzzyViewTruncatesLongEntries(t *testing.T) {
	long := strings.Repeat("x", 200)
	m := newFuzzyModel([]string{long}, "")
	m.width = 20
	view := m.View().Content
	if strings.Contains(view, long) {
		t.Fatal("expected long entry to be truncated")
	}
}
