package main

import (
	"math/rand"
)

const (
	// Hands
	Rock = iota
	Paper
	Scissors

	// Results
	WinP1
	WinP2
	Tie
)

// RandomHand returns a random Hand.
func RandomHand() int {
	return rand.Intn(3)
}

// Play a single game of Rock, Paper, Scissors and return the result.
func Play(p1, p2 int) int {
	switch {
	case p1 == p2:
		return Tie
	case (p1 == Rock && p2 != Paper),
		(p1 == Paper && p2 != Scissors),
		(p1 == Scissors && p2 != Rock):
		return WinP1
	default:
		return WinP2
	}
}
