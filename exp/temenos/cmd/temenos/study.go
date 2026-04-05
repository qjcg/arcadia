package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"temenos/cmd/temenos/templates"
	"temenos/internal/domain"
)

func (app *App) serveStudy(w http.ResponseWriter, r *http.Request, path string) {
	moduleID := path[6:] // remove "/study/" prefix
	if moduleID == "" {
		http.NotFound(w, r)
		return
	}
	ctx := context.Background()

	mod, err := app.db.GetModule(ctx, moduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Module not found: "+moduleID, http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get cards for review (limit to 10)
	cards, err := app.db.ListCardsForReview(ctx, moduleID, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(cards) == 0 {
		templates.StudyComplete(moduleID, mod.Title).Render(r.Context(), w)
		return
	}

	// Get current index from query param
	idxStr := r.URL.Query().Get("idx")
	currentIdx := 0
	if idxStr != "" {
		if parsedIdx, err := strconv.Atoi(idxStr); err == nil && parsedIdx >= 0 && parsedIdx < len(cards) {
			currentIdx = parsedIdx
		}
	}

	// Show current card
	card := cards[currentIdx]
	nextIdx := currentIdx + 1

	// Convert to domain.Card
	domainCard := domain.Card{
		ID:    card.ID,
		Front: card.Front,
		Back:  card.Back,
	}

	templates.StudyCard(moduleID, domainCard, currentIdx, len(cards), nextIdx).Render(r.Context(), w)
}

func (app *App) handleReviewAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	cardID := r.FormValue("card_id")
	quality, err := strconv.Atoi(r.FormValue("quality"))
	if err != nil {
		http.Error(w, "Invalid quality", http.StatusBadRequest)
		return
	}

	// Check if review exists
	existingReview, err := app.db.GetReviewForCard(ctx, cardID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := nowFunc()

	if existingReview == nil {
		// Create new review
		review := domain.Review{
			ID:           "review-" + cardID + "-" + now.Format("20060102150405"),
			CardID:       cardID,
			Quality:      quality,
			Interval:     calculateInterval(quality, 2.5, 0),
			EaseFactor:   2.5,
			ReviewCount:  1,
			NextReviewAt: now.AddDate(0, 0, calculateInterval(quality, 2.5, 0)),
			ReviewedAt:   now,
		}
		if err := app.db.CreateReview(ctx, review); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing review
		existingReview.Quality = quality
		existingReview.EaseFactor = calculateEaseFactor(existingReview.EaseFactor, quality)
		existingReview.Interval = calculateInterval(quality, existingReview.EaseFactor, existingReview.Interval)
		existingReview.ReviewCount++
		existingReview.NextReviewAt = now.AddDate(0, 0, existingReview.Interval)
		existingReview.ReviewedAt = now

		if err := app.db.UpdateReview(ctx, *existingReview); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

// SM-2 algorithm helpers
func calculateEaseFactor(currentEF float64, quality int) float64 {
	// SM-2 formula: EF' = EF + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
	newEF := currentEF + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if newEF < 1.3 {
		return 1.3
	}
	return newEF
}

func calculateInterval(quality int, ef float64, currentInterval int) int {
	if currentInterval == 0 {
		// First review
		switch quality {
		case 0, 1: // Complete failure
			return 1
		case 2: // Hard
			return 1
		case 3: // Good
			return 3
		case 4, 5: // Easy
			return 4
		}
	}

	// Subsequent reviews
	nextInterval := float64(currentInterval) * ef
	if nextInterval < 1 {
		return 1
	}
	return int(nextInterval + 0.5)
}

// nowFunc allows testability
var nowFunc = func() time.Time { return time.Now() }
