// The barr command prints out a system status line.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/qjcg/arcadia/x/barr/blocks"
)

func main() {
	flagSeparator := flag.String("s", "  ", "output field separator")
	flag.Parse()

	// Create a new StatusBar.
	// TODO: Accept config from environment and/or config file as well as
	// defining a default.
	sb := StatusBar{
		Separator: *flagSeparator,
		Stringers: []fmt.Stringer{
			&blocks.WifiData{},
			&blocks.Battery{},
			&blocks.Volume{},
			&blocks.Disk{Dir: "/"},
			// FIXME: Not working! Enable when fixed.
			//&blocks.CryptoCurrency{Pair: "xbtcad"},
			&blocks.LoadAvg{},
			&blocks.DefaultTimeStamp,
		},
	}

	fmt.Println(sb.Get())
}

// StatusBar describes a statusbar.
type StatusBar struct {
	Separator string
	Stringers []fmt.Stringer
}

// Get returns a status string.
func (sb *StatusBar) Get() string {
	var fields []string
	for _, s := range sb.Stringers {
		str := s.String()
		// If a Stringer returns the empty string, it's skipped.
		if str == "" {
			continue
		}
		fields = append(fields, strings.TrimSpace(str))
	}

	return strings.Join(fields, sb.Separator)
}
