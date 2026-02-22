package main

import (
	"testing"
)

func TestRandomHand(t *testing.T) {
	if v := RandomHand(); v > 2 || v < 0 {
		t.Error("expected an int value between 0 and 2, got", v)
	}
}

func TestPlay(t *testing.T) {
	testcases := []struct {
		p1, p2   int
		expected int
	}{
		// tie
		{Rock, Rock, Tie},
		{Paper, Paper, Tie},
		{Scissors, Scissors, Tie},

		// p1 wins
		{Rock, Scissors, WinP1},
		{Paper, Rock, WinP1},
		{Scissors, Paper, WinP1},

		// p2 wins
		{Rock, Paper, WinP2},
		{Paper, Scissors, WinP2},
		{Scissors, Rock, WinP2},
	}

	for _, c := range testcases {
		r := Play(c.p1, c.p2)
		if r != c.expected {
			t.Errorf("expected: %v, got: %v", c.expected, r)
		}
	}
}
