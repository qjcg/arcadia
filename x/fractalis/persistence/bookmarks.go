package persistence

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Bookmark represents a saved fractal location
type Bookmark struct {
	Name                string  `yaml:"name"`
	URL                 string  `yaml:"url,omitempty"`
	FractalType         string  `yaml:"fractal_type"`
	CenterX             float64 `yaml:"center_x"`
	CenterY             float64 `yaml:"center_y"`
	Zoom                float64 `yaml:"zoom"`
	MaxIter             int     `yaml:"max_iter"`
	ColorScheme         string  `yaml:"color_scheme"`
	JuliaCr             float64 `yaml:"julia_cr,omitempty"`
	JuliaCi             float64 `yaml:"julia_ci,omitempty"`
	AutopilotEnabled    bool    `yaml:"autopilot,omitempty"`
	DynamicColorEnabled bool    `yaml:"dynamic_color,omitempty"`
	TransitionMode      string  `yaml:"transition_mode,omitempty"`
}

// BookmarkList holds all bookmarks
type BookmarkList struct {
	Bookmarks []Bookmark `yaml:"bookmarks"`
}

// GetBookmarkPath returns the path to the bookmarks file
func GetBookmarkPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config", "fractals")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "bookmarks.yaml"), nil
}

// LoadBookmarks reads bookmarks from the YAML file
func LoadBookmarks() ([]Bookmark, error) {
	path, err := GetBookmarkPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty list
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Bookmark{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var list BookmarkList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return list.Bookmarks, nil
}

// SaveBookmarks writes bookmarks to the YAML file
func SaveBookmarks(bookmarks []Bookmark) error {
	path, err := GetBookmarkPath()
	if err != nil {
		return err
	}

	list := BookmarkList{Bookmarks: bookmarks}
	data, err := yaml.Marshal(&list)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// GenerateBookmarkName creates a random name in format: adjective_noun
func GenerateBookmarkName() string {
	adjectives := []string{
		"uncharted", "distant", "forgotten", "ancient", "sacred", "secret",
		"mystic", "endless", "infinite", "twilight", "crystal", "phantom",
		"hidden", "silent", "veiled", "ethereal", "serene", "tranquil",
		"arcane", "celestial", "luminous", "radiant", "shadowed", "misty",
		"frosted", "gilded", "bronze", "silver", "crimson", "azure",
		"jade", "amber", "obsidian", "prismatic", "shimmering", "glowing",
		"gleaming", "whispering", "echoing", "wandering",
	}

	journeyNouns := []string{
		"path", "journey", "expedition", "quest", "voyage", "passage",
		"gateway", "portal", "crossing", "frontier", "realm", "domain",
		"territory", "landscape", "horizon", "threshold", "border", "edge",
		"brink", "verge", "summit", "peak", "valley", "canyon",
		"cavern", "hollow", "chamber", "sanctum", "haven", "refuge",
		"shelter", "oasis", "crossroads", "junction", "nexus", "confluence",
		"convergence", "labyrinth", "maze", "corridor",
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	adjective := adjectives[rng.Intn(len(adjectives))]
	noun := journeyNouns[rng.Intn(len(journeyNouns))]

	return adjective + "_" + noun
}
