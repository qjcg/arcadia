package template

name:        "custom"
description: "A custom template for testing"

variables: {
	// Project name
	project_name: string

	// Greeting message
	message?: string | *"Hello, World!"
}
