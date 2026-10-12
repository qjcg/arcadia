//go:build !ebiten

package main

import (
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

func TestRunEbitenEngineStubExplainsBuildTag(t *testing.T) {
	err := runEbitenEngine(persistence.Config{})
	if err == nil {
		t.Fatal("expected error from a build without the ebiten engine")
	}
	if !strings.Contains(err.Error(), "-tags ebiten") {
		t.Errorf("error should mention the ebiten build tag, got: %v", err)
	}
}
