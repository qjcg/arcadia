package persistence

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Bookmark represents a saved fractal location
// All bookmark data is stored in the URL; no separate fields needed
type Bookmark struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
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
	configDir := filepath.Join(homeDir, ".config", "fractalis")

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

	// If file is empty, return empty list
	if len(data) == 0 {
		return []Bookmark{}, nil
	}

	var list BookmarkList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	// Ensure we never return nil - always return an empty slice instead
	// This prevents issues when the YAML has no bookmarks key
	if list.Bookmarks == nil {
		return []Bookmark{}, nil
	}

	return list.Bookmarks, nil
}

// SaveBookmarks writes bookmarks to the YAML file atomically
// to prevent data corruption if the program crashes during write
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

	// Write to a temporary file first, then rename atomically
	// This ensures the bookmarks file is never in a partially-written state
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	// Atomically rename temp file to target file
	return os.Rename(tmpPath, path)
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
