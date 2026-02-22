//go:build integration

package blocks

import (
	"testing"
)

func TestCryptoCurrency(t *testing.T) {
	c := CryptoCurrency{Pair: "xbtcad"}
	t.Log(c.String())
}
