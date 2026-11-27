package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func keyPress(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

// TestFuzzySearchModelIsFocused guards against regressing Ctrl+R search to a
// state where the query input is blurred and silently ignores typing.
func TestFuzzySearchModelIsFocused(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha", "beta"}, "")
	if !m.query.Focused() {
		t.Fatal("fuzzy search query is not focused; typing would be ignored")
	}
	if cmd := m.Init(); cmd == nil {
		t.Fatal("expected a cursor blink command from Init")
	}

	nm, _ := m.Update(keyPress('b'))
	nm, _ = nm.(fuzzyModel).Update(keyPress('e'))
	fm := nm.(fuzzyModel)
	if got := fm.query.Value(); got != "be" {
		t.Fatalf("expected query %q, got %q (typing is not reaching the input)", "be", got)
	}
	if len(fm.filtered) != 1 || fm.filtered[0].entry != "beta" {
		t.Fatalf("expected filtering to narrow to beta, got %+v", fm.filtered)
	}
}

func TestFuzzyUpdateEscape(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b"}, "")
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyEscape}, {Code: 'g', Mod: tea.ModCtrl}} {
		nm, cmd := m.Update(key)
		fm := nm.(fuzzyModel)
		if !fm.done {
			t.Fatal("expected done after escape")
		}
		if fm.result != (fuzzySelection{}) {
			t.Fatalf("expected empty result, got %+v", fm.result)
		}
		if cmd == nil {
			t.Fatal("expected quit command")
		}
	}
}

func TestFuzzyUpdateEnterSelects(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha", "beta"}, "")
	nm, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done after enter")
	}
	if fm.result != (fuzzySelection{line: "alpha", execute: true}) {
		t.Fatalf("expected alpha to execute, got %+v", fm.result)
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestFuzzyUpdateEnterNoResults(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha"}, "zzz")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	fm := nm.(fuzzyModel)
	if !fm.done {
		t.Fatal("expected done")
	}
	if fm.result != (fuzzySelection{}) {
		t.Fatalf("expected empty result with no matches, got %+v", fm.result)
	}
}

func TestFuzzyUpdateCtrlNavigationAndTab(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha", "beta", "gamma"}, "a")
	for _, key := range []rune{'n', 'n', 'p'} {
		next, _ := m.Update(tea.KeyPressMsg{Code: key, Mod: tea.ModCtrl})
		m = next.(fuzzyModel)
	}
	if m.query.Value() != "a" || m.selected != 1 {
		t.Fatalf("navigation changed query or selected wrong row: %q, %d", m.query.Value(), m.selected)
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = next.(fuzzyModel)
	if !m.done || cmd == nil || m.result != (fuzzySelection{line: "beta"}) {
		t.Fatalf("Tab should exit with beta for editing, got %+v", m.result)
	}
}

func TestFuzzyUpdateArrowDown(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b", "c"}, "")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowDownAtEnd(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b"}, "")
	m.selected = 1
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected to stay 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowUp(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b", "c"}, "")
	m.selected = 2
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	fm := nm.(fuzzyModel)
	if fm.selected != 1 {
		t.Fatalf("expected selected 1, got %d", fm.selected)
	}
}

func TestFuzzyUpdateArrowUpAtTop(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b"}, "")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	fm := nm.(fuzzyModel)
	if fm.selected != 0 {
		t.Fatalf("expected selected to stay 0, got %d", fm.selected)
	}
}

func TestFuzzyUpdateHomeClearsQuery(t *testing.T) {
	m := newFuzzySearchModel([]string{"a", "b"}, "a")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	fm := nm.(fuzzyModel)
	if fm.query.Value() != "" {
		t.Fatalf("expected empty query after home, got %q", fm.query.Value())
	}
}

func TestFuzzyUpdateTypingFilters(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha", "beta", "gamma"}, "")
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
	m := newFuzzySearchModel([]string{"alpha", "beta"}, "al")
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	fm := nm.(fuzzyModel)
	if fm.query.Value() != "a" {
		t.Fatalf("expected query 'a' after backspace, got %q", fm.query.Value())
	}
}

func TestFuzzyUpdateWindowSize(t *testing.T) {
	m := newFuzzySearchModel([]string{"a"}, "")
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	fm := nm.(fuzzyModel)
	if fm.width != 100 || fm.height != 30 {
		t.Fatalf("expected (100, 30), got (%d, %d)", fm.width, fm.height)
	}
}

func TestFuzzyUpdateInterrupt(t *testing.T) {
	m := newFuzzySearchModel([]string{"a"}, "")
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
	m := newFuzzySearchModel([]string{"a"}, "")
	m.done = true
	if got := m.View().Content; got != "" {
		t.Fatalf("expected empty view when done, got %q", got)
	}
}

func TestFuzzyViewRendersEntries(t *testing.T) {
	m := newFuzzySearchModel([]string{"alpha", "beta"}, "")
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
	m := newFuzzySearchModel([]string{long}, "")
	m.width = 20
	view := m.View().Content
	if strings.Contains(view, long) {
		t.Fatal("expected long entry to be truncated")
	}
}

func TestFuzzyViewHighlightsAfterTyping(t *testing.T) {
	m := newFuzzySearchModel([]string{"Écho X", "Édit X"}, "")
	for _, r := range "éx" {
		next, _ := m.Update(keyPress(r))
		m = next.(fuzzyModel)
	}
	rows := strings.Split(m.View().Content, "\n")
	for i, row := range rows[:2] {
		base := dimStyle
		if i == 0 {
			base = selectedStyle
			if !strings.Contains(row, base.Render("> ")) {
				t.Fatalf("missing selection marker: %q", row)
			}
		}
		for _, match := range []string{"É", "X"} {
			if !strings.Contains(row, matchStyle(base).Render(match)) {
				t.Errorf("row %d: missing styled match %q in %q", i, match, row)
			}
		}
	}
}
