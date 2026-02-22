package persistence

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildScreenshotContent(t *testing.T) {
	tests := []struct {
		name             string
		fractalType      string
		centerX          float64
		centerY          float64
		zoom             float64
		maxIter          int
		colorScheme      string
		juliaCr          string
		juliaCi          string
		width            int
		height           int
		fractalOutput    string
		checkContains    []string
		checkNotContains []string
	}{
		{
			name:             "basic mandelbrot",
			fractalType:      "mandelbrot",
			centerX:          -0.5,
			centerY:          0.0,
			zoom:             1.0,
			maxIter:          50,
			colorScheme:      "grayscale",
			juliaCr:          "",
			juliaCi:          "",
			width:            80,
			height:           40,
			fractalOutput:    "fractal output here",
			checkContains:    []string{"mandelbrot", "-0.5", "0.0", "1.0", "50", "grayscale", "80x40", "fractal output here"},
			checkNotContains: []string{"Julia"},
		},
		{
			name:             "julia with parameters",
			fractalType:      "julia",
			centerX:          0.0,
			centerY:          0.0,
			zoom:             2.5,
			maxIter:          100,
			colorScheme:      "rainbow",
			juliaCr:          "-0.8",
			juliaCi:          "0.156",
			width:            120,
			height:           60,
			fractalOutput:    "julia output",
			checkContains:    []string{"julia", "2.5", "100", "rainbow", "120x60", "Julia", "-0.8", "0.156", "julia output"},
			checkNotContains: []string{},
		},
		{
			name:             "zoom with scientific notation",
			fractalType:      "mandelbrot",
			centerX:          -0.5,
			centerY:          0.0,
			zoom:             100000.0,
			maxIter:          200,
			colorScheme:      "blue",
			juliaCr:          "",
			juliaCi:          "",
			width:            100,
			height:           50,
			fractalOutput:    "deep zoom output",
			checkContains:    []string{"1.000000e+05", "200", "deep zoom output"},
			checkNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := BuildScreenshotContent(
				tt.fractalType,
				tt.centerX,
				tt.centerY,
				tt.zoom,
				tt.maxIter,
				tt.colorScheme,
				tt.juliaCr,
				tt.juliaCi,
				tt.width,
				tt.height,
				tt.fractalOutput,
			)

			// Verify content is not empty
			if content == "" {
				t.Error("content is empty")
			}

			// Check for required strings
			for _, required := range tt.checkContains {
				if !strings.Contains(content, required) {
					t.Errorf("content missing %q", required)
				}
			}

			// Check for strings that should not be present
			for _, notAllowed := range tt.checkNotContains {
				if strings.Contains(content, notAllowed) {
					t.Errorf("content should not contain %q", notAllowed)
				}
			}
		})
	}
}

func TestBuildScreenshotContentFormat(t *testing.T) {
	// Test that content has proper formatting with separators
	content := BuildScreenshotContent(
		"mandelbrot",
		-0.5, 0.0,
		1.0,
		50,
		"grayscale",
		"", "",
		80, 40,
		"output",
	)

	// Should start with separator line (79 equal signs)
	if !strings.HasPrefix(content, "=") {
		t.Error("content should start with separator")
	}

	// Should contain header section with "Fractal Screenshot"
	if !strings.Contains(content, "Fractal Screenshot") {
		t.Error("content missing header")
	}

	// The separator is: "=" + strings.Repeat("=", 78) = 79 chars
	separator := "=" + strings.Repeat("=", 78)
	separatorCount := strings.Count(content, separator)
	if separatorCount < 2 {
		t.Errorf("expected at least 2 separator lines, got %d", separatorCount)
	}

	// Output should be at the end
	if !strings.HasSuffix(content, "output") {
		t.Error("fractal output should be at end of content")
	}
}

func TestSaveScreenshot(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	content := "test screenshot content"
	filename, err := SaveScreenshot("mandelbrot", content)
	if err != nil {
		t.Fatalf("SaveScreenshot failed: %v", err)
	}

	// Verify filename format
	if !strings.HasPrefix(filename, "mandelbrot_") {
		t.Errorf("filename should start with fractal type, got %q", filename)
	}
	if !strings.HasSuffix(filename, ".txt") {
		t.Errorf("filename should end with .txt, got %q", filename)
	}

	// Verify file exists
	path, err := GetScreenshotPath()
	if err != nil {
		t.Fatalf("GetScreenshotPath failed: %v", err)
	}

	fullPath := filepath.Join(path, filename)
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("screenshot file not created: %v", err)
	}

	// Verify file content
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != content {
		t.Errorf("file content mismatch:\n got: %q\nwant: %q", string(data), content)
	}
}

func TestSaveScreenshotConflict(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	// Save first screenshot
	filename1, err := SaveScreenshot("julia", "content1")
	if err != nil {
		t.Fatalf("first SaveScreenshot failed: %v", err)
	}

	// Save second screenshot with same type (should add counter)
	filename2, err := SaveScreenshot("julia", "content2")
	if err != nil {
		t.Fatalf("second SaveScreenshot failed: %v", err)
	}

	// Filenames should be different
	if filename1 == filename2 {
		t.Errorf("filenames should differ when saving multiple screenshots of same type")
	}

	// Second should have counter
	if !strings.Contains(filename2, "_1.txt") {
		t.Errorf("second filename should contain counter, got %q", filename2)
	}

	// Both files should exist
	path, _ := GetScreenshotPath()
	if _, err := os.Stat(filepath.Join(path, filename1)); err != nil {
		t.Errorf("first file not found: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, filename2)); err != nil {
		t.Errorf("second file not found: %v", err)
	}
}

func TestGetScreenshotPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	defer os.Setenv("HOME", originalPath)

	os.Setenv("HOME", tmpDir)

	path, err := GetScreenshotPath()
	if err != nil {
		t.Fatalf("GetScreenshotPath failed: %v", err)
	}

	// Check path is correct
	expected := filepath.Join(tmpDir, ".config", "fractalis", "screenshots")
	if path != expected {
		t.Errorf("screenshot path: got %q, want %q", path, expected)
	}

	// Check directory was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("screenshot directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("screenshot path is not a directory")
	}
}

func TestScreenshotContentMetadata(t *testing.T) {
	// Test that metadata includes all essential info
	content := BuildScreenshotContent(
		"burningship",
		-0.5, -0.6,
		5.0,
		150,
		"fire",
		"0.2", "0.3",
		200, 100,
		"test output",
	)

	requiredFields := map[string]string{
		"Fractal Type":   "burningship",
		"Center":         "(-0.5000000000, -0.6000000000)",
		"Zoom":           "5.0",
		"Max Iterations": "150",
		"Color Scheme":   "fire",
		"Resolution":     "200x100",
	}

	for fieldName, expectedValue := range requiredFields {
		if !strings.Contains(content, fieldName) {
			t.Errorf("metadata missing field: %s", fieldName)
		}
		if !strings.Contains(content, expectedValue) {
			t.Errorf("metadata missing value: %s", expectedValue)
		}
	}
}

func TestScreenshotJuliaMetadata(t *testing.T) {
	// Non-Julia should not include Julia line
	nonJuliaContent := BuildScreenshotContent(
		"mandelbrot",
		-0.5, 0.0,
		1.0,
		50,
		"grayscale",
		"-0.8", "0.156",
		80, 40,
		"output",
	)

	if strings.Contains(nonJuliaContent, "Julia") {
		t.Error("non-Julia screenshot should not contain Julia parameters")
	}

	// Julia with parameters should include Julia line
	juliaContent := BuildScreenshotContent(
		"julia",
		0.0, 0.0,
		1.0,
		50,
		"blue",
		"-0.8", "0.156",
		80, 40,
		"output",
	)

	if !strings.Contains(juliaContent, "Julia") {
		t.Error("Julia screenshot should contain Julia parameters in metadata")
	}

	// Julia without parameters should not include Julia line
	juliaNoParamsContent := BuildScreenshotContent(
		"julia",
		0.0, 0.0,
		1.0,
		50,
		"blue",
		"", "",
		80, 40,
		"output",
	)

	if strings.Contains(juliaNoParamsContent, "Julia") {
		t.Error("Julia screenshot without parameters should not include Julia line")
	}
}
