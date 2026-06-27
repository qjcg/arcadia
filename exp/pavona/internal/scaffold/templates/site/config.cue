package template

name:        "site"
description: "A static site with Markdown or org-mode content and a custom theme"

variables: {
	site_name: {
		prompt:   "Site name"
		default:  ""
		required: true
		help:     "The name of your site (e.g., My Blog)"
	}
	author: {
		prompt:   "Author name"
		default:  ""
		required: false
	}
	format: {
		prompt:  "Content format"
		default: "markdown"
		choices: ["markdown", "org"]
		required: false
	}
}
