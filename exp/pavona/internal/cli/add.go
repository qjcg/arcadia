package cli

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type AddParams struct {
	Type string `positional:"true" descr:"Component type (handler, feature, tool, page, stream, job, migration)"`
	Name string `positional:"true" descr:"Component name"`
}

func AddCmd() boa.CmdT[AddParams] {
	return boa.CmdT[AddParams]{
		Use:   "add",
		Short: "Add a component to the project",
		Long: `Add a component to the current project.

Types:
  handler   Add an HTTP handler with route
  feature   Add a gherkin feature file with godog step definitions
  tool      Add a CLI subcommand
  page      Add a content page (site projects)
  stream    Add a JetStream stream definition (NATS projects)
  job       Add a background worker
  migration Add a database migration`,
		RunFunc: func(p *AddParams, cmd *cobra.Command, args []string) {
			// TODO: implement add logic per type
			fmt.Fprintf(os.Stderr, "add %s %s: not yet implemented\n", p.Type, p.Name)
			os.Exit(1)
		},
	}
}
