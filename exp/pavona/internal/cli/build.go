package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/pkg/site"
	"github.com/spf13/cobra"
)

type BuildParams struct {
	Dir string `short:"d" descr:"Project directory" default:"." optional:"true"`
}

func BuildCmd() boa.CmdT[BuildParams] {
	return boa.CmdT[BuildParams]{
		Use:   "build",
		Short: "Build a static site to dist/",
		Long:  "Build a static site: renders content/ to dist/ as HTML.",
		RunFunc: func(p *BuildParams, cmd *cobra.Command, args []string) {
			contentDir := filepath.Join(p.Dir, "content")
			outputDir := filepath.Join(p.Dir, "dist")

			if err := site.Build(contentDir, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error building site: %v\n", err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Built site to %s\n", outputDir)
		},
	}
}
