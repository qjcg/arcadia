// An epic Rock, Paper, Scissors game.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mgutz/ansi"
)

const (
	RockAscii     = "0"
	PaperAscii    = "#"
	ScissorsAscii = "X"

	RockLetter     = "R"
	PaperLetter    = "P"
	ScissorsLetter = "S"

	ScissorsSymbol     = "✂"
	ScissorsHandSymbol = "✌"
)

type ScoreBoard struct {
	p1      int
	p2      int
	ties    int
	games   int
	pctP1   float64
	pctP2   float64
	pctTies float64
}

func (s *ScoreBoard) Percentages() {
	s.pctP1 = 100.0 * float64(s.p1) / float64(s.games)
	s.pctP2 = 100.0 * float64(s.p2) / float64(s.games)
	s.pctTies = 100.0 * float64(s.ties) / float64(s.games)
}

func main() {
	HandSymbols := map[int]string{
		Rock:     RockAscii,
		Paper:    PaperAscii,
		Scissors: ScissorsSymbol,
	}

	ngames := flag.Int("n", 100000, "number of games to play")
	live := flag.Bool("l", false, "display live scoreboard updates")
	flag.Parse()

	paleBlue := ansi.ColorFunc("69+b:black")

	// Set up a new scoreboard and play!
	sb := ScoreBoard{}
	for {
		// Exit when ngames have been played
		if sb.games == *ngames {
			// only print scoreboard if no live updates (otherwise redundant)
			if *live {
				fmt.Println()
			} else {
				sb.Percentages()
				fmt.Printf("\rp1: %v (%.3f%%)  p2: %v (%.3f%%)  ties: %v (%.3f%%)  games: %v/%d\n",
					sb.p1, sb.pctP1, sb.p2, sb.pctP2, sb.ties, sb.pctTies, sb.games, *ngames)
			}
			os.Exit(0)
		}

		handP1 := RandomHand()
		handP2 := RandomHand()

		switch Play(handP1, handP2) {
		case Tie:
			sb.ties++
		case WinP1:
			sb.p1++
		case WinP2:
			sb.p2++
		}

		sb.games++
		if *live {
			sb.Percentages()
			fmt.Printf("\r(%.1f%%) %v  %s %s  %v (%.1f%%)  ties: %v (%.1f%%)  games: %v/%d ",
				sb.pctP1, sb.p1,
				paleBlue(HandSymbols[handP1]), paleBlue(HandSymbols[handP2]),
				sb.p2, sb.pctP2,
				sb.ties, sb.pctTies,
				sb.games, *ngames)
		}
	}
}
