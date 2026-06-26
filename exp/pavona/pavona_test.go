package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bolocera/pavona/features/steps"
	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	// Build pavona binary if PAVONA_BIN not set
	if os.Getenv("PAVONA_BIN") == "" {
		bin := filepath.Join(t.TempDir(), "pavona")
		out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
		if err != nil {
			t.Fatalf("building pavona: %v\n%s", err, out)
		}
		t.Setenv("PAVONA_BIN", bin)
	}

	suite := godog.TestSuite{
		Name: "pavona",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			steps.RegisterScaffoldSteps(ctx)
			steps.RegisterSiteSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status")
	}
}
