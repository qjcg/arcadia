package main

import (
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/cli"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

func main() {
	boa.CmdT[cli.TemplateParams]{
		Use:     "pavona",
		Short:   "A cookiecutter-inspired template engine",
		Long:    "Pavona hydrates templates — point it at a template directory\n(or use a built-in), answer a few questions, and get a fully\nrendered project in seconds.",
		Version: getVersion(),
		RunFunc: cli.RunTemplate,
	}.Run()
}
