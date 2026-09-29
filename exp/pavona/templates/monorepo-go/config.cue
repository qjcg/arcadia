package template

name:        "monorepo-go"
description: "A monorepo template using Go"

variables: {
	project_name: string
	go_version:   *"1.27.1" | string
}
