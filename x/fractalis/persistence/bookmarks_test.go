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
			name: "mandelbrot bookmark",
			bookmark: Bookmark{
				Name:        "spiral_region",
				FractalType: "mandelbrot",
				CenterX:     -0.7436,
				CenterY:     0.1314,
				Zoom:        10.0,
				MaxIter:     100,
				ColorScheme: "rainbow",
			},
		},
		{
			name: "julia bookmark with parameters",
			bookmark: Bookmark{
				Name:        "julia_test",
				FractalType: "julia",
				CenterX:     0.0,
				CenterY:     0.0,
				Zoom:        1.0,
				MaxIter:     50,
				ColorScheme: "blue",
				JuliaCr:     -0.8,
				JuliaCi:     0.156,
			},
		},
		{
			name: "bookmark with all optional fields",
			bookmark: Bookmark{
				Name:                "complex_bookmark",
				URL:                 "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t",
				FractalType:         "mandelbrot",
				CenterX:             -0.5,
				CenterY:             0.0,
				Zoom:                1.0,
				MaxIter:             50,
				ColorScheme:         "grayscale",
				JuliaCr:             -0.7,
				JuliaCi:             0.27015,
				AutopilotEnabled:    true,
				DynamicColorEnabled: true,
				TransitionMode:      "fade",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := tt.bookmark
			if bm.Name == "" {
				t.Error("bookmark name is empty")
			}
			if bm.FractalType == "" {
				t.Error("fractal type is empty")
			}
			if bm.MaxIter <= 0 {
				t.Error("MaxIter must be positive")
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
			Name:        "test1",
			FractalType: "mandelbrot",
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
			ColorScheme: "grayscale",
		},
		{
			Name:        "test2",
			FractalType: "julia",
			CenterX:     0.0,
			CenterY:     0.0,
			Zoom:        2.5,
			MaxIter:     75,
			ColorScheme: "blue",
			JuliaCr:     -0.8,
			JuliaCi:     0.156,
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
		if bm.FractalType != bookmarks[i].FractalType {
			t.Errorf("bookmark[%d].FractalType: got %q, want %q", i, bm.FractalType, bookmarks[i].FractalType)
		}
		if bm.CenterX != bookmarks[i].CenterX {
			t.Errorf("bookmark[%d].CenterX: got %v, want %v", i, bm.CenterX, bookmarks[i].CenterX)
		}
		if bm.CenterY != bookmarks[i].CenterY {
			t.Errorf("bookmark[%d].CenterY: got %v, want %v", i, bm.CenterY, bookmarks[i].CenterY)
		}
		if bm.Zoom != bookmarks[i].Zoom {
			t.Errorf("bookmark[%d].Zoom: got %v, want %v", i, bm.Zoom, bookmarks[i].Zoom)
		}
		if bm.MaxIter != bookmarks[i].MaxIter {
			t.Errorf("bookmark[%d].MaxIter: got %d, want %d", i, bm.MaxIter, bookmarks[i].MaxIter)
		}
		if bm.ColorScheme != bookmarks[i].ColorScheme {
			t.Errorf("bookmark[%d].ColorScheme: got %q, want %q", i, bm.ColorScheme, bookmarks[i].ColorScheme)
		}
		if bm.JuliaCr != bookmarks[i].JuliaCr {
			t.Errorf("bookmark[%d].JuliaCr: got %v, want %v", i, bm.JuliaCr, bookmarks[i].JuliaCr)
		}
		if bm.JuliaCi != bookmarks[i].JuliaCi {
			t.Errorf("bookmark[%d].JuliaCi: got %v, want %v", i, bm.JuliaCi, bookmarks[i].JuliaCi)
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
	expected := filepath.Join(tmpDir, ".config", "fractals", "bookmarks.yaml")
	if path != expected {
		t.Errorf("bookmark path: got %q, want %q", path, expected)
	}

	// Check that directory was created
	configDir := filepath.Join(tmpDir, ".config", "fractals")
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
		Name:                "test_roundtrip",
		URL:                 "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t&dynamic_color=t",
		FractalType:         "mandelbrot",
		CenterX:             -0.5,
		CenterY:             0.0,
		Zoom:                1.0,
		MaxIter:             50,
		ColorScheme:         "rainbow",
		JuliaCr:             -0.7,
		JuliaCi:             0.27015,
		AutopilotEnabled:    true,
		DynamicColorEnabled: true,
		TransitionMode:      "fade",
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
