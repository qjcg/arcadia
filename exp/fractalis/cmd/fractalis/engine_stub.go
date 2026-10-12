//go:build !ebiten

package main

import (
	"errors"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// engineChoiceHelp documents the --engine flag choices for this build.
const engineChoiceHelp = "Rendering engine (bubbletea; ebiten not built in)"

// runEbitenEngine reports that this binary was built without the graphical
// engine, keeping fractalis installable on systems without the ebitengine
// native dependencies (X11/ALSA dev libraries, cgo).
func runEbitenEngine(persistence.Config) error {
	return errors.New("the ebiten engine is not compiled into this binary; " +
		"rebuild with `go build -tags ebiten` (see README) or use --engine bubbletea")
}
