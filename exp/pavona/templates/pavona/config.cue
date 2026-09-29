package template

name:        "pavona"
description: "A starter template for creating Pavona templates"

variables: {
	// Template name
	project_name: string

	// Template description
	description?: string | *"A custom template generated with Pavona"
}
