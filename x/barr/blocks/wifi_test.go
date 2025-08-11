//go:build integration

package blocks

import (
	"testing"

	"github.com/mdlayher/wifi"
)

func TestWifiString(t *testing.T) {
	var w WifiData
	err := w.getESSID()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ESSID: %s\n", w.ESSID)
}

func Test_getSSID(t *testing.T) {
	client, err := wifi.New()
	if err != nil {
		t.Fatalf("error creating new wifi client: %v", err)
	}

	interfaces, err := client.Interfaces()
	if err != nil {
		t.Fatalf("error listing network interfaces: %v", err)
	}

	for _, f := range interfaces {
		if f.Name != "" {
			t.Logf("%s", f.Name)
		}
	}
}
