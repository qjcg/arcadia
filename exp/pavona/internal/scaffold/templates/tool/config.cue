package template

name:        "tool"
description: "A Go CLI tool with cobra subcommands and BDD tests"

variables: {
	project_name: {
		prompt:   "Project name"
		default:  ""
		required: true
		help:     "The name of your CLI tool (e.g., gh-deploy)"
	}
	description: {
		prompt:   "Short description"
		default:  "A CLI tool built with Pavona"
		required: false
	}
	version: {
		prompt:   "Initial version"
		default:  "0.1.0"
		required: false
	}
}
