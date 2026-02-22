package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBookmark(t *testing.T) {
	tests := []struct {
		name     string
		bookmark Bookmark
	}{
		{
			name: "simple bookmark with URL",
			bookmark: Bookmark{
				Name: "spiral_region",
				URL:  "fractal://mandelbrot/-0.7436/0.1314/10.0/100/",
			},
		},
		{
			name: "julia bookmark with URL",
			bookmark: Bookmark{
				Name: "julia_test",
				URL:  "fractal://julia/0.0/0.0/1.0/50/?julia_cr=-0.8&julia_ci=0.156",
			},
		},
		{
			name: "bookmark with all query parameters",
			bookmark: Bookmark{
				Name: "complex_bookmark",
				URL:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?color_theme=rainbow&autopilot=t&dynamic_color=t&transition=fade",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := tt.bookmark
			if bm.Name == "" {
				t.Error("bookmark name is empty")
			}
			if bm.URL == "" {
				t.Error("bookmark URL is empty")
			}
			// Verify URL can be parsed
			if _, err := ParseFractalURL(bm.URL); err != nil {
				t.Errorf("bookmark URL cannot be parsed: %v", err)
			}
		})
	}
}

func TestSaveAndLoadBookmarks(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	// Set HOME to temp directory for this test
	os.Setenv("HOME", tmpDir)

	bookmarks := []Bookmark{
		{
			Name: "test1",
			URL:  "fractal://mandelbrot/-0.5/0.0/1.0/50/",
		},
		{
			Name: "test2",
			URL:  "fractal://julia/0.0/0.0/2.5/75/?julia_cr=-0.8&julia_ci=0.156&color_theme=blue",
		},
	}

	// Save bookmarks
	if err := SaveBookmarks(bookmarks); err != nil {
		t.Fatalf("SaveBookmarks failed: %v", err)
	}

	// Load bookmarks back
	loaded, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks failed: %v", err)
	}

	// Verify loaded bookmarks match original
	if len(loaded) != len(bookmarks) {
		t.Errorf("bookmark count mismatch: got %d, want %d", len(loaded), len(bookmarks))
	}

	for i, bm := range loaded {
		if bm.Name != bookmarks[i].Name {
			t.Errorf("bookmark[%d].Name: got %q, want %q", i, bm.Name, bookmarks[i].Name)
		}
		if bm.URL != bookmarks[i].URL {
			t.Errorf("bookmark[%d].URL: got %q, want %q", i, bm.URL, bookmarks[i].URL)
		}
	}
}

func TestLoadBookmarksEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	// Load from non-existent file should return empty slice
	bookmarks, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks should not error on non-existent file: %v", err)
	}

	if len(bookmarks) != 0 {
		t.Errorf("expected empty bookmarks, got %d", len(bookmarks))
	}
}

func TestGetBookmarkPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	path, err := GetBookmarkPath()
	if err != nil {
		t.Fatalf("GetBookmarkPath failed: %v", err)
	}

	// Check path is in home directory
	expected := filepath.Join(tmpDir, ".config", "fractalis", "bookmarks.yaml")
	if path != expected {
		t.Errorf("bookmark path: got %q, want %q", path, expected)
	}

	// Check that directory was created
	configDir := filepath.Join(tmpDir, ".config", "fractalis")
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("config path is not a directory")
	}
}

func TestSaveEmptyBookmarks(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	// Save empty slice
	if err := SaveBookmarks([]Bookmark{}); err != nil {
		t.Fatalf("SaveBookmarks failed: %v", err)
	}

	// Load and verify empty
	bookmarks, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks failed: %v", err)
	}

	if len(bookmarks) != 0 {
		t.Errorf("expected empty bookmarks, got %d", len(bookmarks))
	}
}

func TestBookmarkRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	original := Bookmark{
		Name: "test_roundtrip",
		URL:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?color_theme=rainbow&autopilot=t&dynamic_color=t&transition=fade",
	}

	// Save and load
	if err := SaveBookmarks([]Bookmark{original}); err != nil {
		t.Fatalf("SaveBookmarks failed: %v", err)
	}

	loaded, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(loaded))
	}

	// Deep equality check
	recovered := loaded[0]
	if recovered != original {
		t.Errorf("bookmark roundtrip failed\n got: %+v\nwant: %+v", recovered, original)
	}
}

func TestSaveBookmarksAtomic(t *testing.T) {
	// Test that SaveBookmarks uses atomic write (temp file + rename)
	// This ensures the bookmarks file is never in a partially-written state
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	// Create initial bookmarks
	initial := []Bookmark{
		{Name: "first", URL: "fractal://mandelbrot/-0.5/0.0/1.0/50/"},
	}

	// Save initial bookmarks
	if err := SaveBookmarks(initial); err != nil {
		t.Fatalf("SaveBookmarks failed: %v", err)
	}

	// Verify the bookmarks file exists and temp file does not
	path, _ := GetBookmarkPath()
	tmpPath := path + ".tmp"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("bookmarks file should exist after save")
	}

	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful save")
	}

	// Load and verify content is intact
	loaded, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks failed: %v", err)
	}

	if len(loaded) != 1 || loaded[0].Name != "first" {
		t.Errorf("bookmarks corrupted after save: got %+v", loaded)
	}
}
