package main

import (
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/cli"
	"github.com/spf13/cobra"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

func main() {
	boa.CmdT[struct{}]{
		Use:     "pavona",
		Short:   "A cookiecutter-inspired template engine",
		Long:    "Create projects from built-in or local templates.",
		Version: getVersion(),
		SubCmds: []*cobra.Command{
			cli.NewCmd(),
			cli.ListCmd(),
		},
	}.Run()
}
