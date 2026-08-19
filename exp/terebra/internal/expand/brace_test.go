package expand

import (
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		// No braces
		{"hello", []string{"hello"}},
		{"hello world", []string{"hello world"}},

		// Simple lists
		{"{a,b}", []string{"a", "b"}},
		{"{a,b,c}", []string{"a", "b", "c"}},
		{"{1,2}", []string{"1", "2"}},

		// Prefix/suffix
		{"pre{1,2}post", []string{"pre1post", "pre2post"}},
		{"{a,b}.txt", []string{"a.txt", "b.txt"}},
		{"file{1,2}", []string{"file1", "file2"}},

		// Numeric ranges
		{"{1..5}", []string{"1", "2", "3", "4", "5"}},
		{"{1..3}", []string{"1", "2", "3"}},
		{"{05..08}", []string{"05", "06", "07", "08"}},
		{"{5..1}", []string{"5", "4", "3", "2", "1"}},

		// Numeric ranges with step
		{"{1..10..3}", []string{"1", "4", "7", "10"}},
		{"{10..1..3}", []string{"10", "7", "4", "1"}},

		// Alpha ranges
		{"{a..e}", []string{"a", "b", "c", "d", "e"}},
		{"{e..a}", []string{"e", "d", "c", "b", "a"}},
		{"{A..C}", []string{"A", "B", "C"}},
		{"{a..e..2}", []string{"a", "c", "e"}},

		// Cartesian product (adjacent groups)
		{"{a,b}{1,2}", []string{"a1", "a2", "b1", "b2"}},
		{"{x,y}{1,2}{a,b}", []string{"x1a", "x1b", "x2a", "x2b", "y1a", "y1b", "y2a", "y2b"}},

		// Nested braces
		{"{a,{b,c}}", []string{"a", "b", "c"}},
		{"{a,{b,{c,d}}}", []string{"a", "b", "c", "d"}},

		// Edge cases
		{"{}", []string{"{}"}},
		{"{x}", []string{"x"}},
		{"{1..a}", []string{"{1..a}"}},
		{"{a..1}", []string{"{a..1}"}},

		// Multiple words (not affected by brace expansion)
		{"echo hello", []string{"echo hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Expand(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandSingleElement(t *testing.T) {
	got := Expand("{x}")
	want := []string{"x"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({x}) = %v, want %v", got, want)
	}
}

func TestExpandEmpty(t *testing.T) {
	got := Expand("{}")
	want := []string{"{}"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({}) = %v, want %v", got, want)
	}
}

func TestExpandNoBraces(t *testing.T) {
	got := Expand("hello world")
	want := []string{"hello world"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand(hello world) = %v, want %v", got, want)
	}
}

func TestExpandNested(t *testing.T) {
	got := Expand("{a,{b,c}}")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({a,{b,c}}) = %v, want %v", got, want)
	}
}

func TestExpandCartesianProduct(t *testing.T) {
	got := Expand("{a,b}{1,2}")
	want := []string{"a1", "a2", "b1", "b2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({a,b}{1,2}) = %v, want %v", got, want)
	}
}

func TestExpandNumericRange(t *testing.T) {
	got := Expand("{1..5}")
	want := []string{"1", "2", "3", "4", "5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({1..5}) = %v, want %v", got, want)
	}
}

func TestExpandAlphaRange(t *testing.T) {
	got := Expand("{a..e}")
	want := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({a..e}) = %v, want %v", got, want)
	}
}

func TestExpandReverseRange(t *testing.T) {
	got := Expand("{e..a}")
	want := []string{"e", "d", "c", "b", "a"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({e..a}) = %v, want %v", got, want)
	}
}

func TestExpandZeroPadded(t *testing.T) {
	got := Expand("{01..05}")
	want := []string{"01", "02", "03", "04", "05"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({01..05}) = %v, want %v", got, want)
	}
}

func TestExpandStep(t *testing.T) {
	got := Expand("{1..10..3}")
	want := []string{"1", "4", "7", "10"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({1..10..3}) = %v, want %v", got, want)
	}
}

func TestExpandInvalidRange(t *testing.T) {
	got := Expand("{1..a}")
	want := []string{"{1..a}"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand({1..a}) = %v, want %v", got, want)
	}
}

func TestExpandPrefix(t *testing.T) {
	got := Expand("pre{1,2}post")
	want := []string{"pre1post", "pre2post"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand(pre{1,2}post) = %v, want %v", got, want)
	}
}

func TestExpandMultipleGroups(t *testing.T) {
	got := Expand("f{1,2}.{txt,md}")
	want := []string{"f1.txt", "f1.md", "f2.txt", "f2.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expand(f{1,2}.{txt,md}) = %v, want %v", got, want)
	}
}

func TestFindBracesSimple(t *testing.T) {
	start, end, ok := findBraces("{a,b}")
	if !ok {
		t.Fatal("expected brace group found")
	}
	if start != 0 || end != 4 {
		t.Fatalf("expected (0, 4), got (%d, %d)", start, end)
	}
}

func TestFindBracesPrefixed(t *testing.T) {
	start, end, ok := findBraces("pre{a,b}post")
	if !ok {
		t.Fatal("expected brace group found")
	}
	if start != 3 || end != 7 {
		t.Fatalf("expected (3, 7), got (%d, %d)", start, end)
	}
}

func TestFindBracesNone(t *testing.T) {
	_, _, ok := findBraces("plain")
	if ok {
		t.Fatal("expected no brace group")
	}
}

func TestFindBracesSkippedVar(t *testing.T) {
	_, _, ok := findBraces("${var} plain")
	if ok {
		t.Fatal("expected no brace group (var expansion skipped)")
	}
}

func TestFindBracesNested(t *testing.T) {
	start, end, ok := findBraces("{a,{b,c}}")
	if !ok {
		t.Fatal("expected nested brace group found")
	}
	if start != 0 || end != 8 {
		t.Fatalf("expected (0, 8), got (%d, %d)", start, end)
	}
}

func TestFindBracesNoComma(t *testing.T) {
	start, end, ok := findBraces("{a}")
	if !ok {
		t.Fatal("expected single-element brace group found")
	}
	if start != 0 || end != 2 {
		t.Fatalf("expected (0, 2), got (%d, %d)", start, end)
	}
}

func TestFindBracesUnbalanced(t *testing.T) {
	_, _, ok := findBraces("}{")
	if ok {
		t.Fatal("expected no brace group for unbalanced braces")
	}
}

func TestIsValidBraceContent(t *testing.T) {
	cases := map[string]bool{
		"a,b":      true,
		"1..5":     true,
		"{x}":      true,
		"a..z":     true,
		"a..1":     false,
		"":         false,
		"1..10..2": true,
	}
	for in, want := range cases {
		if got := isValidBraceContent(in); got != want {
			t.Errorf("isValidBraceContent(%q) = %v, want %v", in, got, want)
		}
	}
}
