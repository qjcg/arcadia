package lifecycle

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){})
	os.Exit(m.Run())
}

func TestLifecycle(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			env.Setenv("LIFECYCLE", dir)
			env.Setenv("GOCACHE", env.WorkDir+"/.gocache")
			return nil
		},
	})
}
