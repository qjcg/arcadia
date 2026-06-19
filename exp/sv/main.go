package main

import (
	"os"

	"github.com/qjcg/arcadia/exp/sv/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
