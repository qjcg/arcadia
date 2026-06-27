package template

name:        "lib"
description: "A minimal Go library module with test helpers and BDD scaffold"

variables: {
	project_name: {
		prompt:   "Library name"
		default:  ""
		required: true
		help:     "The name of your Go library (e.g., go-csvstream)"
	}
	description: {
		prompt:   "Short description"
		default:  "A Go library built with Pavona"
		required: false
	}
}
