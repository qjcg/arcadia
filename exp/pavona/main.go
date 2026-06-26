package main

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/bolocera/pavona/internal/cli"
)

func main() {
	boa.CmdT[boa.NoParams]{
		Use:   "pavona",
		Short: "A Go framework that grows with you",
		Long:  "Pavona is a Go framework for building CLI tools, libraries,\nstatic sites, TUIs, web apps, and agents. Named after leaf coral\nof the Pavona genus: layered, branching, and symbiotic.",
		SubCmds: boa.SubCmds(
			cli.NewCmd(),
			cli.AddCmd(),
			cli.RemoveCmd(),
			cli.BuildCmd(),
			cli.ServeCmd(),
		),
	}.Run()
}
