package assets

import "embed"

//go:embed dist/*
var Dist embed.FS
