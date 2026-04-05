package domain

import (
	"time"
)

// CardType represents the type of learning card
type CardType string

const (
	CardTypeFlashcard      CardType = "flashcard"
	CardTypeMultipleChoice CardType = "multiple_choice"
	CardTypeScenario       CardType = "scenario"
	CardTypeInteractive    CardType = "interactive"
)

// Card represents a single learning card
type Card struct {
	ID        string         `json:"id"`
	Type      CardType       `json:"type"`
	ModuleID  string         `json:"module_id"`
	Front     string         `json:"front"`
	Back      string         `json:"back"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Module represents a learning module/pack
type Module struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category,omitempty"`
	Cards       []Card    `json:"cards,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SessionMode represents the learning session mode
type SessionMode string

const (
	SessionModeLearn    SessionMode = "learn"
	SessionModeReview   SessionMode = "review"
	SessionModePractice SessionMode = "practice"
	SessionModeQuiz     SessionMode = "quiz"
)

// Session represents a learning session
type Session struct {
	ID        string      `json:"id"`
	ModuleID  string      `json:"module_id"`
	Mode      SessionMode `json:"mode"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
}

// Review represents a spaced repetition review record
type Review struct {
	ID           string    `json:"id"`
	CardID       string    `json:"card_id"`
	SessionID    string    `json:"session_id,omitempty"`
	Quality      int       `json:"quality"`  // 0-5 for SM-2
	Interval     int       `json:"interval"` // days until next review
	EaseFactor   float64   `json:"ease_factor"`
	ReviewCount  int       `json:"review_count"`
	NextReviewAt time.Time `json:"next_review_at"`
	ReviewedAt   time.Time `json:"reviewed_at"`
}

// NewCard creates a new card with defaults
func NewCard(id string, cardType CardType, moduleID, front, back string) Card {
	now := time.Now()
	return Card{
		ID:        id,
		Type:      cardType,
		ModuleID:  moduleID,
		Front:     front,
		Back:      back,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewModule creates a new module with defaults
func NewModule(id, title, description string) Module {
	now := time.Now()
	return Module{
		ID:          id,
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks if the card is valid
func (c Card) Validate() error {
	if c.ID == "" {
		return ErrInvalidCardID
	}
	if c.ModuleID == "" {
		return ErrInvalidModuleID
	}
	if c.Front == "" {
		return ErrEmptyFront
	}
	if c.Back == "" {
		return ErrEmptyBack
	}
	return nil
}

// Validate checks if the module is valid
func (m Module) Validate() error {
	if m.ID == "" {
		return ErrInvalidModuleID
	}
	if m.Title == "" {
		return ErrEmptyTitle
	}
	return nil
}

// Validate checks if the review is valid
func (r Review) Validate() error {
	if r.ID == "" {
		return ErrInvalidReviewID
	}
	if r.CardID == "" {
		return ErrInvalidCardID
	}
	if r.Quality < 0 || r.Quality > 5 {
		return ErrInvalidQuality
	}
	if r.Interval < 0 {
		return ErrInvalidInterval
	}
	if r.EaseFactor < 1.3 {
		return ErrInvalidEaseFactor
	}
	return nil
}

// Errors for domain validation
var (
	ErrInvalidCardID     = &ValidationError{Field: "id", Message: "card ID cannot be empty"}
	ErrInvalidModuleID   = &ValidationError{Field: "module_id", Message: "module ID cannot be empty"}
	ErrEmptyFront        = &ValidationError{Field: "front", Message: "front content cannot be empty"}
	ErrEmptyBack         = &ValidationError{Field: "back", Message: "back content cannot be empty"}
	ErrEmptyTitle        = &ValidationError{Field: "title", Message: "title cannot be empty"}
	ErrInvalidReviewID   = &ValidationError{Field: "id", Message: "review ID cannot be empty"}
	ErrInvalidQuality    = &ValidationError{Field: "quality", Message: "quality must be between 0 and 5"}
	ErrInvalidInterval   = &ValidationError{Field: "interval", Message: "interval must be non-negative"}
	ErrInvalidEaseFactor = &ValidationError{Field: "ease_factor", Message: "ease factor must be at least 1.3"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
