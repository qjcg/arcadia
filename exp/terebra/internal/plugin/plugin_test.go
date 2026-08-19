package plugin

import (
	"bytes"
	"strings"
	"testing"
)

func TestPluginBuiltinNoArgs(t *testing.T) {
	r := New()
	var out, errb bytes.Buffer
	code := PluginBuiltin(r)(nil, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "expected subcommand") {
		t.Fatalf("expected subcommand error, got %q", errb.String())
	}
}

func TestPluginBuiltinLoadNoPath(t *testing.T) {
	r := New()
	var out, errb bytes.Buffer
	code := PluginBuiltin(r)([]string{"load"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestPluginBuiltinLoadMissingFile(t *testing.T) {
	r := New()
	var out, errb bytes.Buffer
	code := PluginBuiltin(r)([]string{"load", "/nonexistent/plugin.so"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "plugin load:") {
		t.Fatalf("expected load error, got %q", errb.String())
	}
}

func TestPluginBuiltinListEmpty(t *testing.T) {
	r := New()
	var out, errb bytes.Buffer
	code := PluginBuiltin(r)([]string{"list"}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "no plugins loaded") {
		t.Fatalf("expected no plugins message, got %q", out.String())
	}
}

func TestPluginBuiltinUnknownSubcommand(t *testing.T) {
	r := New()
	var out, errb bytes.Buffer
	code := PluginBuiltin(r)([]string{"bogus"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand error, got %q", errb.String())
	}
}

func TestRegistryLoadMissingFile(t *testing.T) {
	r := New()
	if err := r.Load("/nonexistent/plugin.so"); err == nil {
		t.Fatal("expected error loading missing plugin")
	}
}
