//go:build integration
// +build integration

// NOTE: Tests in this file require that a battery be present on
// the test system to work (hence the "integration" build tag).
package blocks

import (
	"testing"
)

func TestBatteryString(t *testing.T) {
	var b Battery
	t.Log(b.String())
}
