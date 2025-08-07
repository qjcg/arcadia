//go:build integration

package script_test

import (
	"context"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

func Test_integrationScripttest(t *testing.T) {
	ctx := context.Background()
	engine := script.NewEngine()

	scripttest.Test(
		t, ctx, engine,
		[]string{
			"PATH=/usr/bin",
		},
		"testdata/ping.txtar")
}
