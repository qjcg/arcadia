package colorthemes

import (
	"fmt"
	"math"
)

// Color scheme types
const (
	ColorGrayscale = "grayscale"
	ColorBlue      = "blue"
	ColorRainbow   = "rainbow"
	ColorFire      = "fire"
	ColorPurple    = "purple"
	ColorGreen     = "green"
	ColorGold      = "gold"
	ColorCyan      = "cyan"
)

// All color schemes for random selection
var AllColorSchemes = []string{
	ColorGrayscale, ColorBlue, ColorRainbow, ColorFire,
	ColorPurple, ColorGreen, ColorGold, ColorCyan,
}

// getColor returns the ANSI color code for a given iteration count and color scheme
func GetColor(iter, maxIter int, scheme string) string {
	return GetColorWithHueShift(iter, maxIter, scheme, 0.0)
}

// getColorWithHueShift returns the ANSI color code with optional hue rotation
func GetColorWithHueShift(iter, maxIter int, scheme string, hueShift float64) string {
	if scheme == ColorGrayscale {
		return ""
	}

	if iter == maxIter {
		// Points in the set are black
		return "\033[38;5;0m"
	}

	// Normalize iteration count to 0-1
	t := float64(iter) / float64(maxIter)

	// Generate base RGB color based on scheme
	var r, g, b int

	switch scheme {
	case ColorBlue:
		// Blue gradient: dark blue to bright blue
		r = int(t * 100)
		g = int(t * 150)
		b = int(100 + t*155)

	case ColorRainbow:
		// Rainbow gradient: red -> yellow -> green -> cyan -> blue
		if t < 0.2 {
			r, g, b = 255, int(t/0.2*255), 0 // Red to yellow
		} else if t < 0.4 {
			r, g, b = int((1.0-(t-0.2)/0.2)*255), 255, 0 // Yellow to green
		} else if t < 0.6 {
			r, g, b = 0, 255, int((t-0.4)/0.2*255) // Green to cyan
		} else if t < 0.8 {
			r, g, b = 0, int((1.0-(t-0.6)/0.2)*255), 255 // Cyan to blue
		} else {
			r, g, b = int((t-0.8)/0.2*128), 0, 255 // Blue to purple
		}

	case ColorFire:
		// Fire gradient: black -> red -> orange -> yellow -> white
		if t < 0.25 {
			r, g, b = int(t/0.25*255), 0, 0 // Black to red
		} else if t < 0.5 {
			r, g, b = 255, int((t-0.25)/0.25*165), 0 // Red to orange
		} else if t < 0.75 {
			r, g, b = 255, int(165+(t-0.5)/0.25*90), int((t-0.5)/0.25*50) // Orange to yellow
		} else {
			base := int((t - 0.75) / 0.25 * 255)
			r, g, b = 255, 255, base // Yellow to white
		}

	case ColorPurple:
		// Purple/magenta gradient
		r = int(128 + t*127)
		g = int(t * 100)
		b = int(200 + t*55)

	case ColorGreen:
		// Green gradient: dark green to bright green
		r = int(t * 100)
		g = int(100 + t*155)
		b = int(t * 100)

	case ColorGold:
		// Gold/amber gradient: brown -> gold -> yellow
		if t < 0.5 {
			r, g, b = int(139+t*116), int(69+t*146), int(19+t*30) // Brown to gold
		} else {
			r, g, b = int(255), int(215+(t-0.5)*80), int(0+(t-0.5)*200) // Gold to yellow
		}

	case ColorCyan:
		// Cyan/aqua gradient: dark cyan to bright cyan
		r = int(t * 100)
		g = int(139 + t*116)
		b = int(139 + t*116)

	default:
		return ""
	}

	// Apply hue shift if enabled
	if hueShift != 0.0 {
		r, g, b = HueShiftColor(r, g, b, hueShift)
	}

	// Convert RGB to ANSI 256 color
	colorIdx := AnsiColorFromRGB(r, g, b)
	return fmt.Sprintf("\033[38;5;%dm", colorIdx)
}

// hueShiftColor applies a hue rotation to an RGB color
// hueShift is in degrees (0-360)
func HueShiftColor(r, g, b int, hueShift float64) (int, int, int) {
	// Convert RGB to HSV
	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var h, s, v float64
	v = max

	if delta < 0.00001 {
		s = 0
		h = 0
	} else {
		s = delta / max

		if rf == max {
			h = 60.0 * (math.Mod((gf-bf)/delta, 6.0))
		} else if gf == max {
			h = 60.0 * (((bf - rf) / delta) + 2.0)
		} else {
			h = 60.0 * (((rf - gf) / delta) + 4.0)
		}

		if h < 0 {
			h += 360.0
		}
	}

	// Apply hue shift
	h = math.Mod(h+hueShift, 360.0)

	// Convert HSV back to RGB
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := v - c

	var rPrime, gPrime, bPrime float64

	if h < 60 {
		rPrime, gPrime, bPrime = c, x, 0
	} else if h < 120 {
		rPrime, gPrime, bPrime = x, c, 0
	} else if h < 180 {
		rPrime, gPrime, bPrime = 0, c, x
	} else if h < 240 {
		rPrime, gPrime, bPrime = 0, x, c
	} else if h < 300 {
		rPrime, gPrime, bPrime = x, 0, c
	} else {
		rPrime, gPrime, bPrime = c, 0, x
	}

	// Convert back to 0-255 range
	rOut := int((rPrime + m) * 255.0)
	gOut := int((gPrime + m) * 255.0)
	bOut := int((bPrime + m) * 255.0)

	return rOut, gOut, bOut
}

// ansiColorFromRGB converts RGB values to the closest ANSI 256-color code
func AnsiColorFromRGB(r, g, b int) int {
	// Use the 216-color cube (colors 16-231)
	// Each component can be 0-5 (6 levels)
	rIdx := int(float64(r) / 255.0 * 5.0)
	gIdx := int(float64(g) / 255.0 * 5.0)
	bIdx := int(float64(b) / 255.0 * 5.0)

	return 16 + (36 * rIdx) + (6 * gIdx) + bIdx
}
