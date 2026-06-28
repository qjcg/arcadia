package template

name:        "agent"
description: "A NATS Agent Protocol service with JetStream"

variables: {
	// Project name (e.g., triagebot)
	project_name: string

	// Short description
	description?: string | *"A NATS agent built with Pavona"
}
