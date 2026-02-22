package main

import (
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"elbereth": main,
	})
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			env.Setenv("GOCACHE", filepath.Join(env.WorkDir, "gocache"))
			env.Setenv("GOPATH", filepath.Join(env.WorkDir, "gopath"))
			return nil
		},
	})
}
