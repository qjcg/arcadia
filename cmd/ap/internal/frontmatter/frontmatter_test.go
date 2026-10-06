package frontmatter

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	doc, err := Parse("---\nname: demo\ndescription: A demo skill.\n---\n\n# Demo\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Meta["name"] != "demo" {
		t.Errorf("name = %v, want demo", doc.Meta["name"])
	}
	if doc.Body != "\n# Demo\n" {
		t.Errorf("body = %q", doc.Body)
	}

	for name, content := range map[string]string{
		"no frontmatter":       "# Demo\n",
		"unclosed frontmatter": "---\nname: demo\n\n# Demo\n",
	} {
		if _, err := Parse(content); err == nil {
			t.Errorf("Parse(%s): want error", name)
		}
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()

	meta, err := Normalize(map[string]any{
		"name":                     "demo",
		"description":              "A demo skill.",
		"disable-model-invocation": true,
		"argument-hint":            "What next?",
		"metadata": map[string]any{
			"credits": map[string]any{
				"skill":  "show-me",
				"author": "Dex Horthy",
			},
		},
	})
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}

	for _, field := range []string{"disable-model-invocation", "argument-hint"} {
		if _, ok := meta[field]; ok {
			t.Errorf("field %q should have been moved into metadata", field)
		}
	}
	metadata, ok := meta["metadata"].(map[string]string)
	if !ok {
		t.Fatalf("metadata = %T, want map[string]string", meta["metadata"])
	}
	want := map[string]string{
		"disable-model-invocation": "true",
		"argument-hint":            "What next?",
		"credits.skill":            "show-me",
		"credits.author":           "Dex Horthy",
	}
	for key, value := range want {
		if metadata[key] != value {
			t.Errorf("metadata[%q] = %q, want %q", key, metadata[key], value)
		}
	}
}

func TestNormalizeRejectsSequences(t *testing.T) {
	t.Parallel()

	if _, err := Normalize(map[string]any{"metadata": map[string]any{"tags": []any{"a"}}}); err == nil {
		t.Error("Normalize: want error for sequence metadata value")
	}
}

func TestRenderDeterministic(t *testing.T) {
	t.Parallel()

	body := "Body stays untouched.\n"
	first, err := Render(map[string]any{"name": "demo", "description": "D."}, body)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	second, err := Render(map[string]any{"description": "D.", "name": "demo"}, body)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if first != second {
		t.Errorf("render not deterministic:\n%s\nvs\n%s", first, second)
	}
	if !strings.HasPrefix(first, "---\nname: demo\ndescription: D.\n---\n") {
		t.Errorf("unexpected render order:\n%s", first)
	}
	if !strings.HasSuffix(first, body) {
		t.Errorf("body not preserved:\n%s", first)
	}
}

func TestNormalizeDocumentRoundTrip(t *testing.T) {
	t.Parallel()

	content := "---\nname: demo\ndescription: \"A demo skill.\"\ndisable-model-invocation: true\n---\n# Demo\n"
	doc, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	out, err := NormalizeDocument(doc)
	if err != nil {
		t.Fatalf("NormalizeDocument: %v", err)
	}
	reparsed, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(normalized): %v", err)
	}
	if reparsed.Meta["name"] != "demo" {
		t.Errorf("name = %v", reparsed.Meta["name"])
	}
	if reparsed.Body != "# Demo\n" {
		t.Errorf("body = %q", reparsed.Body)
	}
	metadata := reparsed.Meta["metadata"].(map[string]any)
	if metadata["disable-model-invocation"] != "true" {
		t.Errorf("metadata = %v", metadata)
	}
}
