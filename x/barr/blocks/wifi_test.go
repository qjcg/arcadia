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
	t.Logf("current SSID: %s", w.SSID)
}
