package main

import (
	"os"

	"github.com/qjcg/arcadia/exp/awesome-lint/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
