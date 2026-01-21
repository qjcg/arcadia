#!/bin/bash
# Test go tools to ensure none are broken due to dependency updates.

go tool cue version &
go tool age --version &
go tool age-inspect --help &
go tool age-keygen --help &
go tool age-plugin-batchpass --help &
go tool litestream version &
go tool svu --version &
go tool editorconfig-checker --version &
go tool lefthook --version &
go tool task --version &
go tool chglog version &
go tool hey --help &
go tool cassowary --version &
go tool sqlc version &
go tool bento --version &
go tool txtar --help &
go tool pkgsite --help &
go tool goimports <<< '' &
go tool govulncheck --version &
go tool gotestsum --version &
go tool helm version &
go tool gofumpt --version &
go tool xurls --version &

wait
