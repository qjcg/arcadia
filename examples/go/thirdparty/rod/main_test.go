//go:build integration

package main_test

import (
	"image"
	_ "image/png"
	"os"
	"path"
	"testing"

	"github.com/go-rod/rod"
)

func Test_MustScreenshot(t *testing.T) {
	page := rod.
		New().
		MustConnect().
		MustPage("https://www.wikipedia.org/")

	outDir := t.TempDir()
	screenshotPath := path.Join(outDir, "wikipedia.png")

	page.
		MustWaitStable().
		MustScreenshot(screenshotPath)

	isPNG, err := isPNGByDecode(screenshotPath)
	if err != nil {
		t.Fatal(err)
	}

	if !isPNG {
		t.Fatalf("screenshot file could not be decoded as a PNG: %s", screenshotPath)
	}

	t.Logf("screenshot decoded as png and saved (%v): %s", isPNG, screenshotPath)
}

func isPNGByDecode(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, format, err := image.DecodeConfig(f)
	if err != nil {
		// not a recognizable image format / corrupt
		return false, nil
	}
	return format == "png", nil
}
