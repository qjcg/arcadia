package main

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"terebra": main,
	})
}

func TestTerebra(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
