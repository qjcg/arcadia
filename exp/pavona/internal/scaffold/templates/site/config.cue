package template

name:        "site"
description: "A static site with Markdown or org-mode content and a custom theme"

variables: {
	// Site name (e.g., My Blog)
	site_name: string

	// Author name
	author?: string | *""

	// Content format
	format?: *"markdown" | "org"
}
