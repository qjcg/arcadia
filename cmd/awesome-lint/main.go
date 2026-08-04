package main

import (
	"os"

	"github.com/qjcg/arcadia/cmd/awesome-lint/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
