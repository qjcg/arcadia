package storage

import (
	"context"
	"testing"
	"time"

	"temenos/internal/domain"
)

func TestStorage_CRUD(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	ctx := context.Background()
	now := time.Now()

	// Create a module
	mod := domain.Module{
		ID:          "poker-basics",
		Title:       "Poker Basics",
		Description: "Learn poker fundamentals",
		Category:    "poker",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.CreateModule(ctx, mod); err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	// Get the module
	fetched, err := db.GetModule(ctx, "poker-basics")
	if err != nil {
		t.Fatalf("failed to get module: %v", err)
	}
	if fetched.Title != "Poker Basics" {
		t.Errorf("expected title %q, got %q", "Poker Basics", fetched.Title)
	}

	// Create cards
	card := domain.Card{
		ID:        "card-1",
		Type:      domain.CardTypeFlashcard,
		ModuleID:  "poker-basics",
		Front:     "What is a flush?",
		Back:      "Five cards of the same suit",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateCard(ctx, card); err != nil {
		t.Fatalf("failed to create card: %v", err)
	}

	// List cards by module
	cards, err := db.ListCardsByModule(ctx, "poker-basics")
	if err != nil {
		t.Fatalf("failed to list cards: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}

	// Get card
	fetchedCard, err := db.GetCard(ctx, "card-1")
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}
	if fetchedCard.Front != "What is a flush?" {
		t.Errorf("expected front %q, got %q", "What is a flush?", fetchedCard.Front)
	}

	// List cards for review (card with no review record should show up)
	newCards, err := db.ListCardsForReview(ctx, "poker-basics", 10)
	if err != nil {
		t.Fatalf("failed to list cards for review: %v", err)
	}
	if len(newCards) != 1 {
		t.Errorf("expected 1 new card for review, got %d", len(newCards))
	}

	// Create a review
	review := domain.Review{
		ID:           "review-1",
		CardID:       "card-1",
		Quality:      4,
		Interval:     1,
		EaseFactor:   2.5,
		ReviewCount:  1,
		NextReviewAt: now.Add(24 * time.Hour),
		ReviewedAt:   now,
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	// Get review for card
	fetchedReview, err := db.GetReviewForCard(ctx, "card-1")
	if err != nil {
		t.Fatalf("failed to get review: %v", err)
	}
	if fetchedReview.Quality != 4 {
		t.Errorf("expected quality %d, got %d", 4, fetchedReview.Quality)
	}

	// Update review
	review.Quality = 5
	review.Interval = 3
	review.EaseFactor = 2.6
	review.ReviewCount = 2
	review.NextReviewAt = now.Add(3 * 24 * time.Hour)
	review.ReviewedAt = now
	if err := db.UpdateReview(ctx, review); err != nil {
		t.Fatalf("failed to update review: %v", err)
	}

	// Create a session
	session := domain.Session{
		ID:        "session-1",
		ModuleID:  "poker-basics",
		Mode:      domain.SessionModeReview,
		StartedAt: now,
	}
	if err := db.CreateSession(ctx, session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Get session
	fetchedSession, err := db.GetSession(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if fetchedSession.Mode != domain.SessionModeReview {
		t.Errorf("expected mode %q, got %q", domain.SessionModeReview, fetchedSession.Mode)
	}

	// Count cards
	count, err := db.CountCardsByModule(ctx, "poker-basics")
	if err != nil {
		t.Fatalf("failed to count cards: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count %d, got %d", 1, count)
	}

	// List modules
	modules, err := db.ListModules(ctx)
	if err != nil {
		t.Fatalf("failed to list modules: %v", err)
	}
	if len(modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(modules))
	}
}
