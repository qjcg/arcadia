//go:build packages
// +build packages

//go:generate go build
//go:generate upx dude
//go:generate mandown -install . dude.1.md
//go:generate holo-build --force --format=rpm holo.toml
//go:generate holo-build --force --format=pacman holo.toml
//go:generate rm dude
package main
