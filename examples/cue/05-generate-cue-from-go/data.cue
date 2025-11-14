package main

import (
	appsv1 "k8s.io/api/apps/v1"
)

deployments: [Name=string]: {
	appsv1.#Deployment

	apiVersion: "apps/v1"
	kind:       "Deployment"

	metadata: labels: {
		app:           *Name | string
		deploymentEnv: *"dev" | "uat" | "prod"
	}

	let Labels = metadata.labels

	spec: {
		selector:
			matchLabels: Labels
		template: metadata: {
			labels: Labels
		}
	}
}

deployments: alpine: {
	spec: template: spec: {
		containers: [{
			name:  "myalpine"
			image: "alpine:3"
		}]
	}
}
