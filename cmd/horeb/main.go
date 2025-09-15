// A CLI tool for generating random sequences of characters from various Unicode blocks.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"log/slog"
	"time"
	"runtime/debug"

	"github.com/qjcg/arcadia/cmd/horeb/internal/blocks"
	"github.com/samber/mo"
)

const (
	description = "horeb: Speaking in tongues via stdout"
)

// config contains our application config.
type config struct {
	debug                  *bool
	listBlocks             *bool
	listBlocksWithContents *bool
	nChars                 *int
	ofs                    *string
	stream                 *bool
	streamDelay            *time.Duration
	version                *bool
	blocks                 []string
}

func getConf(w io.Writer, args []string) mo.Result[*config] {
	var err error

	fs := flag.NewFlagSet("horeb", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(w, "\n%s\n\n", description)
		fs.PrintDefaults()
		fmt.Fprintln(w)
	}

	conf := config{
		debug:                  fs.Bool("d", false, "print debug log messages"),
		listBlocks:             fs.Bool("l", false, "list all blocks"),
		listBlocksWithContents: fs.Bool("L", false, "list all blocks along with their contents"),
		nChars:                 fs.Int("n", 30, "number of runes to generate"),
		ofs:                    fs.String("o", " ", "output field separator"),
		stream:                 fs.Bool("s", false, "generate an endless stream of runes"),
		streamDelay:            fs.Duration("D", time.Millisecond*30, "stream delay"),
		version:                fs.Bool("v", false, "print version"),
	}
	if err = fs.Parse(args); err != nil {
		return mo.Err[*config](err)
	}

	conf.blocks = []string{"all"}
	if fs.NArg() > 0 {
		conf.blocks = fs.Args()
	}

	slog.Debug("configuration parsed from command line args", "conf", conf, "args", fs.Args())

	return mo.Ok(&conf)
}

func main() {
	conf, err := getConf(os.Stderr, os.Args[1:]).Get()
	if err != nil {
		slog.Error("error getting flags", "error", err)
		os.Exit(1)
	}

	if *conf.version {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			slog.Error("error reading build info")
			os.Exit(1)
		}

		fmt.Println(buildInfo.Main.Version)
		os.Exit(0)
	}

	// special value means all blocks
	if conf.blocks[0] == "all" {
		// remove "all" value after use
		conf.blocks = conf.blocks[:0]
		for k := range blocks.Blocks {
			conf.blocks = append(conf.blocks, k)
		}
	}

	switch {
	case *conf.listBlocks:
		blocks.ListBlocks(os.Stdout)
	case *conf.listBlocksWithContents:
		blocks.ListBlocksWithContents(os.Stdout)

	// PrintRandom or StreamRandom from a _single_ block.
	case len(conf.blocks) == 1:
		b, ok := blocks.Blocks[conf.blocks[0]]
		if !ok {
			err := errors.New("unknown block")
			slog.Error("Unknown block", "error", err, "block", conf.blocks[0])
			os.Exit(1)
		}

		if *conf.stream {
			b.StreamRandom(os.Stdout, *conf.streamDelay, *conf.ofs)
		} else {
			b.PrintRandom(os.Stdout, *conf.nChars, *conf.ofs)
		}

	// Print a RandomRune or stream from two or more blocks.
	case len(conf.blocks) > 1:
		bm := map[string]blocks.UnicodeBlock{}
		for _, b := range conf.blocks {
			val, ok := blocks.Blocks[b]
			if !ok {
				slog.Warn("Unknown block", "block", b)
				continue
			}
			bm[b] = val
		}
		if len(bm) > 0 {
			defer fmt.Println()
			if *conf.stream {
				ticker := time.NewTicker(*conf.streamDelay)
				for range ticker.C {

					block, err := blocks.RandomBlock(bm)
					if err != nil {
						slog.Error("error getting random block", "error", err)
						os.Exit(1)
					}
					fmt.Printf("%c%s", block.RandomRune(), *conf.ofs)
				}
			} else {
				for i := 0; i < *conf.nChars; i++ {
					block, err := blocks.RandomBlock(bm)
					if err != nil {
						slog.Error("error getting random block", "error", err)
						os.Exit(1)
					}
					fmt.Printf("%c%s", block.RandomRune(), *conf.ofs)
				}
			}
		}
	}
}
