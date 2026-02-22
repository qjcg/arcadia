package persistence

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
)

// FractalURLParams holds parsed parameters from a fractal:// URL
type FractalURLParams struct {
	Mode                URLMode
	FractalType         string
	CenterX             float64
	CenterY             float64
	Zoom                float64
	MaxIter             int
	ColorTheme          string
	Transition          string
	Engine              string
	AutopilotEnabled    bool
	DynamicColorEnabled bool
	JuliaCr             float64
	JuliaCi             float64
}

// ParseFractalURL parses a fractal:// URL and returns the parameters
func ParseFractalURL(urlString string) (FractalURLParams, error) {
	// Manual parsing because url.Parse treats first path segment as host
	// when there's no slashes after scheme
	if !strings.HasPrefix(urlString, "fractal://") {
		return FractalURLParams{}, fmt.Errorf("invalid scheme: expected 'fractal://', got '%s'", strings.Split(urlString, "://")[0])
	}

	// Remove scheme
	rest := strings.TrimPrefix(urlString, "fractal://")

	// Split path and query
	var pathStr, queryStr string
	if idx := strings.IndexByte(rest, '?'); idx >= 0 {
		pathStr = rest[:idx]
		queryStr = rest[idx+1:]
	} else {
		pathStr = rest
		queryStr = ""
	}

	params := FractalURLParams{
		// Defaults
		ColorTheme:          color.ColorGrayscale,
		Transition:          "none",
		Engine:              EngineBubbleTea,
		AutopilotEnabled:    false,
		DynamicColorEnabled: false,
		JuliaCr:             -0.7,
		JuliaCi:             0.27015,
	}

	// Parse path components
	pathParts := strings.Split(strings.Trim(pathStr, "/"), "/")

	// Filter out empty parts
	var filteredParts []string
	for _, part := range pathParts {
		if part != "" {
			filteredParts = append(filteredParts, part)
		}
	}
	pathParts = filteredParts

	// Check if this is random mode
	if len(pathParts) > 0 && pathParts[0] == "random" {
		params.Mode = ModeRandom
		params.FractalType = FractalMandelbrot // Default, will be overridden by random
	} else if len(pathParts) >= 5 {
		// Standard mode: fractal_name/x/y/z/iter
		params.Mode = ModeStandard

		// Parse fractal type
		params.FractalType = pathParts[0]
		if !IsValidFractalType(params.FractalType) {
			return FractalURLParams{}, fmt.Errorf("invalid fractal type: %s", params.FractalType)
		}

		// Parse coordinates and zoom
		centerX, err := strconv.ParseFloat(pathParts[1], 64)
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid center_x: %s", pathParts[1])
		}
		params.CenterX = centerX

		centerY, err := strconv.ParseFloat(pathParts[2], 64)
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid center_y: %s", pathParts[2])
		}
		params.CenterY = centerY

		zoom, err := strconv.ParseFloat(pathParts[3], 64)
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid zoom: %s", pathParts[3])
		}
		if zoom <= 0 {
			return FractalURLParams{}, fmt.Errorf("zoom must be positive, got %f", zoom)
		}
		params.Zoom = zoom

		maxIter, err := strconv.Atoi(pathParts[4])
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid max_iter: %s", pathParts[4])
		}
		if maxIter <= 0 {
			return FractalURLParams{}, fmt.Errorf("max_iter must be positive, got %d", maxIter)
		}
		params.MaxIter = maxIter
	} else if len(pathParts) > 0 {
		return FractalURLParams{}, fmt.Errorf("invalid URL format: need fractal_name/x/y/z/iter/ or random/")
	}

	// Parse query parameters
	queryParams, err := url.ParseQuery(queryStr)
	if err != nil {
		return FractalURLParams{}, fmt.Errorf("invalid query string: %w", err)
	}

	if colorTheme := queryParams.Get("color_theme"); colorTheme != "" {
		if !IsValidColorTheme(colorTheme) {
			return FractalURLParams{}, fmt.Errorf("invalid color_theme: %s", colorTheme)
		}
		params.ColorTheme = colorTheme
	}

	if transition := queryParams.Get("transition"); transition != "" {
		if !IsValidTransition(transition) {
			return FractalURLParams{}, fmt.Errorf("invalid transition: %s", transition)
		}
		params.Transition = transition
	}

	if engine := queryParams.Get("engine"); engine != "" {
		switch strings.ToLower(engine) {
		case EngineBubbleTea, EngineEbiten:
			params.Engine = strings.ToLower(engine)
		default:
			return FractalURLParams{}, fmt.Errorf("invalid engine: %s", engine)
		}
	}

	// Parse boolean variables: can be specified without value (defaults to true) or with "t"/"f"
	if _, ok := queryParams["autopilot"]; ok {
		if parsedAutopilot, err := ParseBooleanParam(queryParams, "autopilot"); err != nil {
			return FractalURLParams{}, err
		} else {
			params.AutopilotEnabled = parsedAutopilot
		}
	}

	if _, ok := queryParams["dynamic_color"]; ok {
		if parsedDynamicColor, err := ParseBooleanParam(queryParams, "dynamic_color"); err != nil {
			return FractalURLParams{}, err
		} else {
			params.DynamicColorEnabled = parsedDynamicColor
		}
	}

	if juliaCr := queryParams.Get("julia_cr"); juliaCr != "" {
		cr, err := strconv.ParseFloat(juliaCr, 64)
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid julia_cr: %s", juliaCr)
		}
		params.JuliaCr = cr
	}

	if juliaCi := queryParams.Get("julia_ci"); juliaCi != "" {
		ci, err := strconv.ParseFloat(juliaCi, 64)
		if err != nil {
			return FractalURLParams{}, fmt.Errorf("invalid julia_ci: %s", juliaCi)
		}
		params.JuliaCi = ci
	}

	return params, nil
}

// ConfigToFractalURL generates a fractal:// URL from a Config and runtime state
func ConfigToFractalURL(config Config, autopilot, dynamicColor bool, transitionMode int) string {
	// Build the base URL path
	basePath := fmt.Sprintf("fractal://%s/%.10g/%.10g/%.10g/%d/",
		config.FractalType,
		config.CenterX,
		config.CenterY,
		config.Zoom,
		config.MaxIter,
	)

	// Build query parameters
	queryParams := url.Values{}

	// Color theme
	if config.ColorScheme != "" && config.ColorScheme != color.ColorGrayscale {
		queryParams.Set("color_theme", config.ColorScheme)
	}

	// Transition
	if transitionMode > 0 {
		transitionName := TransitionModeToString(transitionMode)
		if transitionName != "none" {
			queryParams.Set("transition", transitionName)
		}
	}

	// Engine
	if config.Engine != "" && config.Engine != EngineBubbleTea {
		queryParams.Set("engine", config.Engine)
	}

	// Autopilot
	if autopilot {
		queryParams.Set("autopilot", "t")
	}

	// Dynamic color
	if dynamicColor {
		queryParams.Set("dynamic_color", "t")
	}

	// Julia parameters
	if config.FractalType == FractalJulia {
		if config.JuliaCr != 0 {
			queryParams.Set("julia_cr", strconv.FormatFloat(config.JuliaCr, 'g', -1, 64))
		}
		if config.JuliaCi != 0 {
			queryParams.Set("julia_ci", strconv.FormatFloat(config.JuliaCi, 'g', -1, 64))
		}
	}

	// Append query string
	urlStr := basePath
	if queryString := queryParams.Encode(); queryString != "" {
		urlStr += "?" + queryString
	}

	return urlStr
}

// RandomToFractalURL generates a random URL with query parameters
func RandomToFractalURL(colorTheme string, autopilot, dynamicColor bool, transitionMode int) string {
	queryParams := url.Values{}

	if colorTheme != "" && colorTheme != color.ColorGrayscale {
		queryParams.Set("color_theme", colorTheme)
	}

	if transitionMode > 0 {
		transitionName := TransitionModeToString(transitionMode)
		if transitionName != "none" {
			queryParams.Set("transition", transitionName)
		}
	}

	if autopilot {
		queryParams.Set("autopilot", "t")
	}

	if dynamicColor {
		queryParams.Set("dynamic_color", "t")
	}

	urlStr := "fractal://random/"
	if queryString := queryParams.Encode(); queryString != "" {
		urlStr += "?" + queryString
	}

	return urlStr
}

// ValidateFractalURL checks if a URL is valid without fully parsing it
func ValidateFractalURL(urlString string) error {
	_, err := ParseFractalURL(urlString)
	return err
}

// Helper functions

// ParseBooleanParam parses a boolean query parameter.
// If the parameter is present but empty (e.g., ?autopilot), it returns true.
// If the parameter has a value, it must be "t" (true) or "f" (false).
// Returns (value, bool, error) where bool indicates if parameter was found.
func ParseBooleanParam(queryParams url.Values, key string) (bool, error) {
	values, ok := queryParams[key]
	if !ok {
		return false, nil
	}

	// If present but no value or empty value, treat as true
	if len(values) == 0 {
		return true, nil
	}

	val := strings.ToLower(values[0])
	switch val {
	case "", "t":
		return true, nil
	case "f":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s value: %s (use 't' for true or 'f' for false, or specify without value for true)", key, values[0])
	}
}

func IsValidFractalType(fractalType string) bool {
	validTypes := map[string]bool{
		FractalMandelbrot:    true,
		FractalJulia:         true,
		FractalBurningShip:   true,
		FractalTricorn:       true,
		FractalMultibrot3:    true,
		FractalMultibrot4:    true,
		FractalCeltic:        true,
		FractalPerpendicular: true,
		FractalMultibrot5:    true,
		FractalManhattan:     true,
		FractalNewton:        true,
		FractalMandelbulb:    true,
		FractalMandelbox:     true,
	}
	return validTypes[fractalType]
}

func IsValidColorTheme(theme string) bool {
	validThemes := map[string]bool{
		color.ColorGrayscale: true,
		color.ColorBlue:      true,
		color.ColorRainbow:   true,
		color.ColorFire:      true,
		color.ColorPurple:    true,
		color.ColorGreen:     true,
		color.ColorGold:      true,
		color.ColorCyan:      true,
	}
	return validThemes[theme]
}

func IsValidTransition(transition string) bool {
	validTransitions := map[string]bool{
		"none":         true,
		"fade":         true,
		"zoomout":      true,
		"zoom_out":     true,
		"rotate":       true,
		"breakthrough": true,
	}
	return validTransitions[transition]
}

func TransitionModeToString(mode int) string {
	switch mode {
	case TransitionNone:
		return "none"
	case TransitionFade:
		return "fade"
	case TransitionZoomOutIn:
		return "zoomout"
	case TransitionRotate:
		return "rotate"
	case TransitionBreakthrough:
		return "breakthrough"
	default:
		return "none"
	}
}

func StringToTransitionMode(s string) int {
	switch strings.ToLower(s) {
	case "fade":
		return TransitionFade
	case "zoomout", "zoom_out":
		return TransitionZoomOutIn
	case "rotate":
		return TransitionRotate
	case "breakthrough":
		return TransitionBreakthrough
	default:
		return TransitionNone
	}
}
