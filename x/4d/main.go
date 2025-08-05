// A simple CLI stopwatch.
package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

const usage = `A simple CLI stopwatch.
Usage: 4d [DURATION]

4d		display elapsed time
4d 15m		countdown 15 minutes
4d 3h2m1s	countdown 3 hours, 2 minutes, 1 second

Ctrl-C exits.`

// fmtDuration returns a HH:MM:SS string representation of a duration.
func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("\r%02d:%02d:%02d ",
		int(d.Hours())%60,
		int(d.Minutes())%60,
		int(d.Seconds())%60,
	)
}

// Countdown outputs time remaining relative to a starting duration at the
// interval specified by the provided ticker.
func Countdown(w io.Writer, ticker *time.Ticker, d time.Duration) {
	start := time.Now()
	// One second is added here to make the starting countdown time
	// correspond to the provided duration.
	end := start.Add(d + time.Second)

	// This for loop style will run one iteration immediately, unlike "for
	// range ticker.C", which waits one tick before printing anything.
	for ; true; <-ticker.C {
		remaining := time.Until(end)
		// Comparing to 1.5s avoids counting the "00:00:00" second. We
		// added 1s to the "end" time above, so this ensures the
		// duration of our countdown is accurate.
		if remaining >= time.Millisecond*1500 {
			fmt.Fprint(w, fmtDuration(remaining))
		} else {
			fmt.Fprintln(w)
			break
		}
	}
}

// Elapsed outputs the duration since *start* at the interval
// specified by the provided ticker.
func Elapsed(w io.Writer, ticker *time.Ticker, start time.Time) {
	for ; true; <-ticker.C {
		fmt.Fprint(w, fmtDuration(time.Since(start)))
	}
}

func main() {
	var countdown time.Duration

	// Parse duration if provided.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "-help", "--help", "help":
			fmt.Printf("\n%s\n\n", usage)
			os.Exit(0)
		}

		d, err := time.ParseDuration(os.Args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		countdown = d
	}

	ticker := time.NewTicker(time.Second)
	if countdown >= time.Second {
		Countdown(os.Stdout, ticker, countdown)
	} else {
		Elapsed(os.Stdout, ticker, time.Now())
	}
}
