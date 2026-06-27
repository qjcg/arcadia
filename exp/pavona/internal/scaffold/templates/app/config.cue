package template

name:        "app"
description: "A full-stack web app with templ, HTMX, SQLite, and Tailwind/DaisyUI"

variables: {
	// Project name (e.g., acmecorp)
	project_name: string

	// Short description
	description?: string | *"A web app built with Pavona"

	// Include a demo URL shortener?
	demo?: "true" | *"false"
}
