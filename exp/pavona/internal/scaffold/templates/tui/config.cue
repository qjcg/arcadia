package template

name:        "tui"
description: "A terminal UI app with bubbletea components"

variables: {
	project_name: {
		prompt:   "Project name"
		default:  ""
		required: true
		help:     "The name of your TUI app (e.g., chatmonitor)"
	}
	description: {
		prompt:   "Short description"
		default:  "A TUI app built with Pavona"
		required: false
	}
}
