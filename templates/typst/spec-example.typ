#import "spec.typ": spec

#show: spec.with(
  title: [Example Software Specification],
  author: [Crush AI],
)

= Introduction
Template fixed: title page, TOC, content, end page.

== Functional Requirements
- Passkeys authentication
- Casbin RBAC
- SQLite with sqlc

== Architecture
- Go backend (internal packages)
- Templ + HTMX + Tailwind frontend

= Deployment
- Taskfile.yml builds
- Docker services

= Foo
#lorem(200)
