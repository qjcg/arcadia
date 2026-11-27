package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalLinks(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	source := "# Café\n# Café\n\n[duplicate](#caf%C3%A9-1)\n\n[valid][target]\n\n[target]: other%20page.md#explicit\n\n[missing anchor](#missing)\n\n[missing file](gone.md)\n\n`[example](ignored.md)`\n\n[external](https://example.com/)\n"
	for name, body := range map[string]string{"README.md": source, "other page.md": `<a id="explicit"></a>`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if status := run([]string{"README.md"}, false, &out); status != 1 {
		t.Fatalf("status = %d; %s", status, &out)
	}
	if lines := strings.Split(strings.TrimSpace(out.String()), "\n"); len(lines) != 2 || !strings.Contains(lines[0], `anchor "missing"`) || !strings.Contains(lines[1], `"gone.md"`) {
		t.Fatalf("expected only the missing file and anchor diagnostics, got:\n%s", &out)
	}
}

func TestExternalResponses(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []int
		want     string
	}{
		{"recover from rate limit", []int{429, 200}, "ok"},
		{"missing page", []int{404}, "broken"},
		{"gone page", []int{410}, "broken"},
		{"access blocked", []int{403}, "unverified"},
		{"server unavailable", []int{503, 503, 503}, "unverified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan int, len(tc.statuses)+1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				index := len(requests)
				requests <- index
				if index >= len(tc.statuses) {
					http.Error(w, "unexpected retry", 400)
					return
				}
				w.WriteHeader(tc.statuses[index])
				_, _ = fmt.Fprintln(w, "response")
			}))
			defer srv.Close()
			got, detail := checkHTTP(srv.Client(), srv.URL, func(time.Duration) {})
			if got != tc.want || len(requests) != len(tc.statuses) {
				t.Fatalf("got %s (%s) after %d requests; want %s after %d", got, detail, len(requests), tc.want, len(tc.statuses))
			}
		})
	}
}
