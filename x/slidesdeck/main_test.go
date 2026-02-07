package main

import (
	"os"
	"testing"

	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/cli"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"slidesdeck": mainFunc,
	}))
}

func mainFunc() int {
	if err := cli.Execute(); err != nil {
		return 1
	}
	return 0
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
