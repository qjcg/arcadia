//go:build integration
// +build integration

// NOTE: Tests in this file require that an audio card be present on
// the test system to work (hence the "integration" build tag).
package blocks

import (
	"testing"
)

func TestVolumeString(t *testing.T) {
	v := Volume{}
	t.Log(v.String())
}
