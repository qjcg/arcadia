package color

import (
	"slices"
	"testing"
)

func TestGetColor_Grayscale(t *testing.T) {
	color := GetColor(10, 100, ColorGrayscale)
	if color != "" {
		t.Errorf("Expected empty string for grayscale, got '%s'", color)
	}
}

func TestGetColor_MaxIterIsBlack(t *testing.T) {
	color := GetColor(100, 100, ColorBlue)
	expected := "\033[38;5;0m"
	if color != expected {
		t.Errorf("Expected black for max iteration, got '%s'", color)
	}
}

func TestGetColor_RainbowScheme(t *testing.T) {
	color := GetColor(50, 100, ColorRainbow)
	if color == "" {
		t.Error("Expected non-empty color for rainbow scheme")
	}
	if len(color) < 8 {
		t.Errorf("Expected ANSI color code, got '%s'", color)
	}
}

func TestGetColor_BlueScheme(t *testing.T) {
	color := GetColor(25, 100, ColorBlue)
	if color == "" {
		t.Error("Expected non-empty color for blue scheme")
	}
}

func TestGetColor_InvalidScheme(t *testing.T) {
	color := GetColor(10, 100, "invalid-scheme")
	if color != "" {
		t.Errorf("Expected empty string for invalid scheme, got '%s'", color)
	}
}

func TestGetColorWithHueShift_NoShift(t *testing.T) {
	color1 := GetColor(50, 100, ColorBlue)
	color2 := GetColorWithHueShift(50, 100, ColorBlue, 0.0)
	if color1 != color2 {
		t.Errorf("Expected same color with zero shift, got '%s' vs '%s'", color1, color2)
	}
}

func TestGetColorWithHueShift_WithHueShift(t *testing.T) {
	color1 := GetColor(50, 100, ColorBlue)
	color2 := GetColorWithHueShift(50, 100, ColorBlue, 60.0)
	if color1 == color2 {
		t.Error("Expected different colors with hue shift")
	}
}

func TestAllColorSchemes_ContainsExpected(t *testing.T) {
	expected := []string{ColorGrayscale, ColorBlue, ColorRainbow, ColorFire, ColorPurple, ColorGreen, ColorGold, ColorCyan}
	if len(AllColorSchemes) != len(expected) {
		t.Errorf("Expected %d schemes, got %d", len(expected), len(AllColorSchemes))
	}
	for _, scheme := range expected {
		found := slices.Contains(AllColorSchemes, scheme)
		if !found {
			t.Errorf("Expected scheme '%s' in AllColorSchemes", scheme)
		}
	}
}
