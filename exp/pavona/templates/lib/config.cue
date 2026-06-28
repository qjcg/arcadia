package template

name:        "lib"
description: "A minimal Go library module with test helpers and BDD scaffold"

variables: {
	// Library name (e.g., go-csvstream)
	project_name: string

	// Short description
	description?: string | *"A Go library built with Pavona"
}
