package ui

import "github.com/qjcg/arcadia/x/fractals/persistence"

// BookmarkState holds bookmark management UI state
type BookmarkState struct {
	ShowBookmarks     bool                   // Show bookmark list
	Bookmarks         []persistence.Bookmark // Loaded bookmarks
	BookmarkCursor    int                    // Selected bookmark in list
	SavingBookmark    bool                   // Prompting for bookmark name
	BookmarkInput     string                 // User input for bookmark name
	SuggestedBookmark string                 // Auto-generated bookmark name suggestion
}

// UIState holds all UI overlay state
type UIState struct {
	ShowHelp bool
	Bookmark BookmarkState
}

// NewUIState creates initialized UI state
func NewUIState() UIState {
	return UIState{
		Bookmark: BookmarkState{},
	}
}
