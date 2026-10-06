package cli

import (
	"fmt"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/ap/internal/spec"
	"github.com/qjcg/arcadia/exp/ap/internal/sync"
	"github.com/spf13/cobra"
)

type syncParams struct {
	Specs string `descr:"Sync specs directory (default: agents/plugins/specs)" optional:"true"`
	Out   string `descr:"Output directory for generated packages (default: agents/plugins)" optional:"true"`
}

func createSyncCmd() *cobra.Command {
	return boa.CmdT[syncParams]{
		Use:   "sync",
		Short: "Regenerate plugin packages from their sync specs",
		Long: "Regenerate every plugin package under the output directory from its sync spec.\n\n" +
			"Each spec pins a skill source (a git repo at a ref, or a local tree) and the " +
			"manifest fields for plugin.json. Skills are flattened into skills/<name>/, " +
			"frontmatter is normalized to Agent Skills form, and output is deterministic: " +
			"re-running sync on the same inputs produces byte-identical trees.",
		RunFuncE: func(p *syncParams, cmd *cobra.Command, _ []string) error {
			return runSyncCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runSyncCmd(p *syncParams, cmd *cobra.Command) error {
	specsDir := p.Specs
	if specsDir == "" {
		specsDir = "agents/plugins/specs"
	}
	outDir := p.Out
	if outDir == "" {
		outDir = "agents/plugins"
	}

	specs, err := spec.Load(specsDir)
	if err != nil {
		return err
	}
	if len(specs) == 0 {
		return fmt.Errorf("no sync specs found in %s", specsDir)
	}
	for _, s := range specs {
		fmt.Fprintf(cmd.OutOrStdout(), "syncing %s from %s\n", s.Name, sync.String(s))
	}
	roots, err := sync.Run(specs, ".", outDir)
	if err != nil {
		return err
	}
	for _, root := range roots {
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", root)
	}
	return nil
}
