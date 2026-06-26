package cli

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type RemoveParams struct {
	Type string `positional:"true" descr:"Component type (handler, feature, tool, page, stream, job, migration)"`
	Name string `positional:"true" descr:"Component name"`
}

func RemoveCmd() boa.CmdT[RemoveParams] {
	return boa.CmdT[RemoveParams]{
		Use:   "remove",
		Short: "Remove a component from the project",
		Long: `Remove a previously added component.

Types:
  handler   Remove an HTTP handler and its route
  feature   Remove a gherkin feature file and its step definitions
  tool      Remove a CLI subcommand
  page      Remove a content page
  stream    Remove a JetStream stream definition
  job       Remove a background worker
  migration Remove a database migration`,
		RunFunc: func(p *RemoveParams, cmd *cobra.Command, args []string) {
			// TODO: implement remove logic per type
			fmt.Fprintf(os.Stderr, "remove %s %s: not yet implemented\n", p.Type, p.Name)
			os.Exit(1)
		},
	}
}
