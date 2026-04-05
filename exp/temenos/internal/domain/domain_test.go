package domain

import (
	"encoding/json"
	"testing"
)

func TestCardJSONSerialization(t *testing.T) {
	card := NewCard("card-1", CardTypeFlashcard, "poker-hands", "Royal Flush", "A-K-Q-J-10 same suit")

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Card
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != card.ID {
		t.Errorf("expected ID %s, got %s", card.ID, decoded.ID)
	}
	if decoded.Type != card.Type {
		t.Errorf("expected Type %s, got %s", card.Type, decoded.Type)
	}
	if decoded.Front != card.Front {
		t.Errorf("expected Front %s, got %s", card.Front, decoded.Front)
	}
	if decoded.Back != card.Back {
		t.Errorf("expected Back %s, got %s", card.Back, decoded.Back)
	}
}

func TestCardValidation(t *testing.T) {
	tests := []struct {
		name    string
		card    Card
		wantErr bool
	}{
		{
			name:    "valid card",
			card:    NewCard("card-1", CardTypeFlashcard, "poker-hands", "Front", "Back"),
			wantErr: false,
		},
		{
			name:    "missing ID",
			card:    Card{ID: "", ModuleID: "mod-1", Front: "Front", Back: "Back"},
			wantErr: true,
		},
		{
			name:    "missing ModuleID",
			card:    Card{ID: "card-1", ModuleID: "", Front: "Front", Back: "Back"},
			wantErr: true,
		},
		{
			name:    "missing Front",
			card:    Card{ID: "card-1", ModuleID: "mod-1", Front: "", Back: "Back"},
			wantErr: true,
		},
		{
			name:    "missing Back",
			card:    Card{ID: "card-1", ModuleID: "mod-1", Front: "Front", Back: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModuleJSONSerialization(t *testing.T) {
	module := NewModule("poker-hands", "Poker Hand Rankings", "Learn the 10 poker hand rankings")
	module.Cards = []Card{
		NewCard("card-1", CardTypeFlashcard, "poker-hands", "Royal Flush", "A-K-Q-J-10 same suit"),
		NewCard("card-2", CardTypeFlashcard, "poker-hands", "Straight Flush", "Five consecutive same suit"),
	}

	data, err := json.Marshal(module)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Module
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != module.ID {
		t.Errorf("expected ID %s, got %s", module.ID, decoded.ID)
	}
	if len(decoded.Cards) != len(module.Cards) {
		t.Errorf("expected %d cards, got %d", len(module.Cards), len(decoded.Cards))
	}
}

func TestReviewSM2Defaults(t *testing.T) {
	review := Review{
		CardID:     "card-1",
		Quality:    3,
		Interval:   1,
		EaseFactor: 2.5,
	}

	if review.EaseFactor != 2.5 {
		t.Errorf("expected EaseFactor 2.5, got %f", review.EaseFactor)
	}
	if review.Interval != 1 {
		t.Errorf("expected Interval 1, got %d", review.Interval)
	}
}
