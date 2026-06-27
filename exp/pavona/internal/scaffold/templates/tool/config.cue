package template

name:        "tool"
description: "A Go CLI tool with cobra subcommands and BDD tests"

variables: {
	// Project name (e.g., gh-deploy)
	project_name: string

	// Short description
	description?: string | *"A CLI tool built with Pavona"

	// Initial version
	version?: string | *"0.1.0"
}
