package template

name:        "app"
description: "A full-stack web app with templ, HTMX, SQLite, and Tailwind/DaisyUI"

variables: {
	project_name: {
		prompt:   "Project name"
		default:  ""
		required: true
		help:     "The name of your web app (e.g., acmecorp)"
	}
	description: {
		prompt:   "Short description"
		default:  "A web app built with Pavona"
		required: false
	}
	demo: {
		prompt:  "Include demo URL shortener?"
		default: "false"
		choices: ["true", "false"]
		required: false
	}
}
