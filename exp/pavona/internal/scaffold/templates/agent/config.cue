package template

name:        "agent"
description: "A NATS Agent Protocol service with JetStream"

variables: {
	project_name: {
		prompt:   "Project name"
		default:  ""
		required: true
		help:     "The name of your agent service (e.g., triagebot)"
	}
	description: {
		prompt:   "Short description"
		default:  "A NATS agent built with Pavona"
		required: false
	}
}
