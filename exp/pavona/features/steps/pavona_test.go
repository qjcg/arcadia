package steps

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs all Gherkin feature scenarios using godog.
func TestFeatures(t *testing.T) {
	// Build the binary once for the entire test suite
	state := &PavonaState{
		tmpDir: t.TempDir(),
	}

	if err := state.buildBinary(); err != nil {
		t.Fatalf("building pavona binary: %v", err)
	}

	suite := godog.TestSuite{
		Name: "pavona",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			// Reset state before each scenario
			ctx.Before(func(ctx2 context.Context, sc *godog.Scenario) (context.Context, error) {
				state.outputDir = filepath.Join(state.tmpDir, sc.Name)
				state.reset()
				return ctx2, nil
			})

			RegisterListSteps(ctx, state)
			RegisterHydrateSteps(ctx, state)
			RegisterCustomSteps(ctx, state)
			RegisterErrorSteps(ctx, state)
			RegisterVersionSteps(ctx, state)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{FeaturesDir},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status from godog test suite")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
