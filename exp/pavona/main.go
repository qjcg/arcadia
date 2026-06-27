package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/cli"
	"github.com/qjcg/arcadia/exp/pavona/internal/scaffold"
	"github.com/spf13/cobra"
)

func getBuiltinTemplateNames() []string {
	list := scaffold.ListBuiltin()
	names := make([]string, len(list))
	for i, t := range list {
		names[i] = t.Name
	}
	return names
}

// listTemplateDirs returns subdirectories under dir that contain config.cue,
// prefixed with prefix (e.g. "./").
func listTemplateDirs(dir, prefix string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "config.cue")); err == nil {
				dirs = append(dirs, prefix+e.Name())
			}
		}
	}
	return dirs
}

func completeTemplateFlag(cmd *cobra.Command, args []string, toComplete string) []string {
	builtins := getBuiltinTemplateNames()
	candidates := make([]string, 0, len(builtins)+10)
	candidates = append(candidates, builtins...)

	// If toComplete contains a path separator, complete from that directory.
	// Otherwise, complete from the current directory with "./" prefix so the
	// user can tab-complete relative paths.
	if idx := strings.LastIndexAny(toComplete, "/\\"); idx >= 0 {
		baseDir := toComplete[:idx+1]
		absDir := filepath.Dir(toComplete)
		if absDir == "." {
			absDir = "."
		}
		if dirs := listTemplateDirs(absDir, baseDir); dirs != nil {
			candidates = append(candidates, dirs...)
		}
	} else {
		if dirs := listTemplateDirs(".", "./"); dirs != nil {
			candidates = append(candidates, dirs...)
		}
	}

	return candidates
}

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
		InitFuncCtx: func(ctx *boa.HookContext, params *cli.TemplateParams, cmd *cobra.Command) error {
			ctx.GetParam(&params.Template).SetAlternativesFunc(completeTemplateFlag)
			return nil
		},
		RunFunc: cli.RunTemplate,
	}.Run()
}
