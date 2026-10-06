// Command ap generates and validates Agent Plugin packages under agents/plugins.
package main

import (
	"os"

	"github.com/qjcg/arcadia/cmd/ap/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
