package template

name:        "site"
description: "A static site with Markdown or org-mode content and a custom theme"

variables: {
	// Project name (e.g., my-blog)
	project_name: string

	// Author name
	author?: string | *""

	// Content format
	format?: *"markdown" | "org"
}
