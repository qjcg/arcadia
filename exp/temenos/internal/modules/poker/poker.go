package poker

// HandRankings returns flashcard content for all 10 poker hand rankings
func HandRankings() []Flashcard {
	return []Flashcard{
		{
			Front: "Royal Flush",
			Back:  "A, K, Q, J, 10 all of the same suit. The highest hand in poker. Example: A♥ K♥ Q♥ J♥ 10♥",
		},
		{
			Front: "Straight Flush",
			Back:  "Five consecutive cards of the same suit. Example: 9♦ 8♦ 7♦ 6♦ 5♦",
		},
		{
			Front: "Four of a Kind (Quads)",
			Back:  "Four cards of the same rank. Example: A♠ A♥ A♦ A♣ K♠",
		},
		{
			Front: "Full House",
			Back:  "Three of a kind plus a pair. Example: K♠ K♥ K♦ 8♣ 8♥",
		},
		{
			Front: "Flush",
			Back:  "Five cards of the same suit, not in sequence. Example: A♣ K♣ 9♣ 7♣ 4♣",
		},
		{
			Front: "Straight",
			Back:  "Five consecutive cards of different suits. Example: 10♠ 9♥ 8♦ 7♣ 6♠",
		},
		{
			Front: "Three of a Kind (Trips)",
			Back:  "Three cards of the same rank. Example: Q♠ Q♥ Q♦ A♣ K♥",
		},
		{
			Front: "Two Pair",
			Back:  "Two different pairs. Example: J♠ J♦ 9♥ 9♣ A♦",
		},
		{
			Front: "One Pair",
			Back:  "Two cards of the same rank. Example: 10♠ 10♦ A♣ K♥ Q♥",
		},
		{
			Front: "High Card",
			Back:  "No poker hand made. The highest card wins. Example: A♠ K♦ 9♥ 7♣ 4♦",
		},
	}
}

// Flashcard represents a simple front/back card
type Flashcard struct {
	Front string
	Back  string
}
