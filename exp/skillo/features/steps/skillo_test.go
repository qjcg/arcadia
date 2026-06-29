package steps

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	testDir := t.TempDir()
	state := &State{testDir: testDir}
	if err := state.buildBinary(); err != nil {
		t.Fatalf("building binary: %v", err)
	}

	suite := godog.TestSuite{
		Name: "skillo",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				state.resetState()
				return ctx, nil
			})
			RegisterAllSteps(ctx, state)
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
