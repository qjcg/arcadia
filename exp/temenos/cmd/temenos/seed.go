package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"temenos/internal/domain"
	"temenos/internal/modules/poker"
	"temenos/internal/storage"
)

func seed(db *storage.DB) error {
	ctx := context.Background()

	// Check if we already have modules
	modules, err := db.ListModules(ctx)
	if err != nil {
		return err
	}
	if len(modules) > 0 {
		log.Println("Database already seeded, skipping")
		return nil
	}

	now := time.Now()

	// Create Poker Basics module
	pokerBasics := domain.Module{
		ID:          "poker-basics",
		Title:       "Poker Basics",
		Description: "Learn the fundamentals of poker",
		Category:    "poker",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.CreateModule(ctx, pokerBasics); err != nil {
		return err
	}

	// Add cards from poker module
	handRankings := poker.HandRankings()
	cards := make([]domain.Card, len(handRankings))
	for i, hr := range handRankings {
		cards[i] = domain.Card{
			ID:        fmt.Sprintf("poker-%d", i+1),
			Type:      domain.CardTypeFlashcard,
			ModuleID:  "poker-basics",
			Front:     hr.Front,
			Back:      hr.Back,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	for _, card := range cards {
		if err := db.CreateCard(ctx, card); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d cards in %s module", len(cards), pokerBasics.Title)
	return nil
}
