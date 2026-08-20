package expand

import (
	"os"
	"os/user"
	"testing"
)

func TestTilde(t *testing.T) {
	home := "/home/testuser"
	os.Setenv("HOME", home)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	tests := []struct {
		name string
		word Word
		want string
	}{
		{"bare tilde", Word{Value: "~"}, home},
		{"tilde slash", Word{Value: "~/foo"}, home + "/foo"},
		{"tilde mid-word", Word{Value: "a~b"}, "a~b"},
		{"tilde not at start", Word{Value: "x~"}, "x~"},
		{"single quoted", Word{Value: "~", Mask: []bool{true}}, "~"},
		{"double quoted", Word{Value: "~/foo", Mask: []bool{true, true, true, true, true, true}}, "~/foo"},
		{"escaped", Word{Value: "~", Mask: []bool{true}}, "~"},
		{"tilde plus out of scope", Word{Value: "~+"}, "~+"},
		{"tilde minus out of scope", Word{Value: "~-"}, "~-"},
		{"unknown user", Word{Value: "~nosuchuser_xyz"}, "~nosuchuser_xyz"},
		{"unknown user slash", Word{Value: "~nosuchuser_xyz/foo"}, "~nosuchuser_xyz/foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tilde(tt.word)
			if got.Value != tt.want {
				t.Fatalf("Tilde(%q) = %q, want %q", tt.word.Value, got.Value, tt.want)
			}
		})
	}
}

func TestTildeUnsetHome(t *testing.T) {
	os.Unsetenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", "/home/testuser") })

	got := Tilde(Word{Value: "~"})
	if got.Value != "~" {
		t.Fatalf("Tilde with unset HOME = %q, want literal ~", got.Value)
	}
}

func TestTildeUser(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	if u.HomeDir == "" {
		t.Skip("current user has no home dir")
	}

	got := Tilde(Word{Value: "~" + u.Username + "/foo"})
	want := u.HomeDir + "/foo"
	if got.Value != want {
		t.Fatalf("Tilde(~%s/foo) = %q, want %q", u.Username, got.Value, want)
	}
}

func TestTildeHomeMaskQuoted(t *testing.T) {
	// A home path containing a glob char must be marked quoted so it is not
	// re-globbed. Simulate by checking the mask length matches the value.
	home := "/home/te*st"
	os.Setenv("HOME", home)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	got := Tilde(Word{Value: "~/x"})
	if got.Value != home+"/x" {
		t.Fatalf("Tilde = %q, want %q", got.Value, home+"/x")
	}
	if len(got.Mask) != len(got.Value) {
		t.Fatalf("mask length %d != value length %d", len(got.Mask), len(got.Value))
	}
	// The home bytes must be quoted (true).
	for i := 0; i < len(home); i++ {
		if !got.Mask[i] {
			t.Fatalf("home byte %d not marked quoted", i)
		}
	}
	// The rest must keep its original (unquoted) mask.
	for i := len(home); i < len(got.Value); i++ {
		if got.Mask[i] {
			t.Fatalf("rest byte %d unexpectedly quoted", i)
		}
	}
}
