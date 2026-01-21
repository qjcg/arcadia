package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetScreenshotPath returns the directory path for screenshots
func GetScreenshotPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	screenshotDir := filepath.Join(homeDir, ".config", "fractals", "screenshots")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		return "", err
	}

	return screenshotDir, nil
}

// SaveScreenshot saves the current fractal view to a text file
// fractalType: name of the fractal type
// content: the fractal output with metadata to save
// Returns the filename that was written
func SaveScreenshot(fractalType string, content string) (string, error) {
	dir, err := GetScreenshotPath()
	if err != nil {
		return "", err
	}

	// Generate filename with timestamp and fractal type
	now := time.Now()
	timestamp := now.Format("2006-01-02_150405")
	filename := fmt.Sprintf("%s_%s.txt", fractalType, timestamp)
	fullPath := filepath.Join(dir, filename)

	// Check if file exists and add counter if needed
	counter := 1
	for {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("%s_%s_%d.txt", fractalType, timestamp, counter)
		fullPath = filepath.Join(dir, filename)
		counter++
	}

	// Write to file
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", err
	}

	return filename, nil
}

// BuildScreenshotContent creates the full screenshot content with metadata
// Parameters: fractalType, center coordinates, zoom level, maxIter, colorScheme,
// julia parameters (only used if fractalType is "julia"), resolution, and fractal output
func BuildScreenshotContent(fractalType string, centerX, centerY, zoom float64,
	maxIter int, colorScheme, juliaCr, juliaCi string, width, height int, fractalOutput string,
) string {
	now := time.Now()

	var metadata strings.Builder
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n")
	metadata.WriteString(fmt.Sprintf("Fractal Screenshot - %s\n", now.Format("2006-01-02 15:04:05")))
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n")
	metadata.WriteString(fmt.Sprintf("Fractal Type: %s\n", fractalType))
	metadata.WriteString(fmt.Sprintf("Center: (%.10f, %.10f)\n", centerX, centerY))

	// Format zoom display
	var zoomStr string
	if zoom >= 10000.0 {
		zoomStr = fmt.Sprintf("%.6e", zoom)
	} else {
		zoomStr = fmt.Sprintf("%.6f", zoom)
	}
	metadata.WriteString(fmt.Sprintf("Zoom: %sx\n", zoomStr))
	metadata.WriteString(fmt.Sprintf("Max Iterations: %d\n", maxIter))
	metadata.WriteString(fmt.Sprintf("Color Scheme: %s\n", colorScheme))

	if juliaCr != "" && juliaCi != "" && fractalType == "julia" {
		metadata.WriteString(fmt.Sprintf("Julia Parameters: c = %s + %si\n", juliaCr, juliaCi))
	}

	metadata.WriteString(fmt.Sprintf("Resolution: %dx%d\n", width, height))
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n\n")

	// Combine metadata and fractal output
	return metadata.String() + fractalOutput
}
