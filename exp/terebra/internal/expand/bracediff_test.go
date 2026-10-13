package expand_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	mexpand "mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"

	"github.com/qjcg/arcadia/exp/terebra/internal/expand"
)

// TestBraceDiffAgainstMvdanSh is the differential harness from issue
// #13: the documented brace-expansion behavior matrix, run through
// terebra's custom expander and mvdan/sh's SplitBraces + BracesSeq.
// mvdanWant is nil when both must agree; a non-nil value records an
// accepted divergence with its decision in gap.
func TestBraceDiffAgainstMvdanSh(t *testing.T) {
	cases := []struct {
		name      string
		word      string
		want      []string
		mvdanWant []string
		gap       string
	}{
		{
			name: "list",
			word: "{a,b,c}",
			want: []string{"a", "b", "c"},
		},
		{
			name: "list_with_prefix_suffix",
			word: "pre{a,b}post",
			want: []string{"preapost", "prebpost"},
		},
		{
			name: "numeric_range",
			word: "{1..5}",
			want: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "zero_padded_range",
			word: "{01..05}",
			want: []string{"01", "02", "03", "04", "05"},
		},
		{
			name: "alpha_range",
			word: "{a..f}",
			want: []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			name: "uppercase_range",
			word: "{A..Z}",
			want: []string{
				"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
				"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
				"Y", "Z",
			},
		},
		{
			name: "stepped_range",
			word: "{1..10..2}",
			want: []string{"1", "3", "5", "7", "9"},
		},
		{
			name: "stepped_alpha_range",
			word: "{a..z..2}",
			want: []string{
				"a", "c", "e", "g", "i", "k", "m", "o", "q", "s", "u", "w",
				"y",
			},
		},
		{
			name: "reverse_numeric_range",
			word: "{5..1}",
			want: []string{"5", "4", "3", "2", "1"},
		},
		{
			name: "nested_list",
			word: "{a,{b,c}}",
			want: []string{"a", "b", "c"},
		},
		{
			name: "cartesian_lists",
			word: "{a,b}{1,2}",
			want: []string{"a1", "a2", "b1", "b2"},
		},
		{
			name: "cartesian_range_list",
			word: "{1..2}{a,b}",
			want: []string{"1a", "1b", "2a", "2b"},
		},
		{
			name: "empty_element",
			word: "{a,}",
			want: []string{"a", ""},
		},
		{
			name:      "single_element_literal",
			word:      "{a}",
			want:      []string{"a"},
			mvdanWant: []string{"{a}"},
			gap: "bash keeps {a} literal: braces expand only with a " +
				"comma or .. range. Per ADR 0004's superset rule bash " +
				"wins; terebra's expansion here is dropped with the " +
				"custom expander.",
		},
		{
			name: "empty_literal",
			word: "{}",
			want: []string{"{}"},
		},
		{
			name: "unclosed_literal",
			word: "x{a,b",
			want: []string{"x{a,b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mvdanWant := tc.mvdanWant
			if mvdanWant == nil {
				mvdanWant = tc.want
			}
			if got := expand.Expand(tc.word); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("terebra Expand(%q) = %q, want %q",
					tc.word, got, tc.want)
			}
			if got := mvdanBraces(t, tc.word); !reflect.DeepEqual(got, mvdanWant) {
				t.Errorf("mvdan braces(%q) = %q, want %q",
					tc.word, got, mvdanWant)
			}
			if tc.gap != "" {
				t.Logf("accepted divergence: %s", tc.gap)
			}
		})
	}
}

// mvdanBraces expands one word the way the post-migration engine will:
// parse as a bash word, split brace parts, then stream BracesSeq.
func mvdanBraces(t *testing.T, word string) []string {
	t.Helper()
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader("echo "+word), "")
	if err != nil {
		t.Fatalf("mvdan parse %q: %v", word, err)
	}
	call := prog.Stmts[0].Cmd.(*syntax.CallExpr)
	wd := call.Args[1]
	syntax.SplitBraces(wd)
	cfg := &mexpand.Config{}
	var got []string
	for w, err := range mexpand.BracesSeq(cfg, wd) {
		if err != nil {
			t.Fatalf("mvdan braces %q: %v", word, err)
		}
		var buf bytes.Buffer
		if err := syntax.NewPrinter().Print(&buf, w); err != nil {
			t.Fatalf("mvdan print %q: %v", word, err)
		}
		got = append(got, buf.String())
	}
	return got
}
