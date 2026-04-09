package main

import (
	"runtime/debug"
)

// Version returns this tool's semantic version number from build info.
func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}
