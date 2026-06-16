package main

import (
	"fmt"
	"os"

	"github.com/qjcg/arcadia/exp/slidesdeck/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
