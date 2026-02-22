//go:build integration

package blocks

import (
	"testing"
)

func TestWifiString(t *testing.T) {
	var w WifiData
	err := w.getSSID()
	if err != nil {
		t.Fatal(err)
	}
	t.Attr("SSID", w.SSID)
}
