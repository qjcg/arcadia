package main

import (
	"os"

	"github.com/qjcg/arcadia/cmd/sv/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
