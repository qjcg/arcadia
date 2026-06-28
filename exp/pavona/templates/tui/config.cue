package template

name:        "tui"
description: "A terminal UI app with bubbletea components"

variables: {
	// Project name (e.g., chatmonitor)
	project_name: string

	// Short description
	description?: string | *"A TUI app built with Pavona"
}
