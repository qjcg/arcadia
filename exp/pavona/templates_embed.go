package main

import (
	"embed"

	"github.com/qjcg/arcadia/exp/pavona/internal/scaffold"
)

//go:embed templates/tool
var toolTmpls embed.FS

//go:embed templates/lib
var libTmpls embed.FS

//go:embed templates/site
var siteTmpls embed.FS

//go:embed templates/tui
var tuiTmpls embed.FS

//go:embed templates/app
var appTmpls embed.FS

//go:embed templates/agent
var agentTmpls embed.FS

func init() {
	scaffold.RegisterBuiltin(
		"tool", "A Go CLI tool with cobra subcommands and BDD tests",
		toolTmpls, "templates/tool",
		map[string]string{
			".gitignore": "/{{.project_name}}\n",
		},
		[]string{"features"},
	)
	scaffold.RegisterBuiltin(
		"lib", "A minimal Go library module with test helpers",
		libTmpls, "templates/lib",
		map[string]string{
			".gitignore": "/bin/\n",
		},
		[]string{"features"},
	)
	scaffold.RegisterBuiltin(
		"site", "A static site with Markdown or org-mode content",
		siteTmpls, "templates/site",
		nil,
		nil,
	)
	scaffold.RegisterBuiltin(
		"tui", "A terminal UI app using bubbletea",
		tuiTmpls, "templates/tui",
		map[string]string{
			".gitignore": "/bin/\n",
		},
		[]string{"features", "views", "models"},
	)
	scaffold.RegisterBuiltin(
		"app", "A full-stack web app with templ, SQLite, HTMX, Tailwind",
		appTmpls, "templates/app",
		map[string]string{
			".gitignore": "/{{.project_name}}\n",
		},
		[]string{"features", "features/steps", "handlers", "static", "internal", "demo"},
	)
	scaffold.RegisterBuiltin(
		"agent", "A NATS Agent Protocol service",
		agentTmpls, "templates/agent",
		map[string]string{
			".gitignore": "/bin/\n",
		},
		[]string{"features", "agent", "nats"},
	)
}
