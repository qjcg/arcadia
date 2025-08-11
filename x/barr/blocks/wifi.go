package blocks

import (
	"errors"
	"fmt"

	"github.com/mdlayher/wifi"
)

type WifiData struct {
	SSID string
}

// String implements the fmt.Stringer interface.
func (w *WifiData) String() string {
	err := w.getSSID()
	if err != nil || w.SSID == "" {
		return ""
	}

	return w.SSID
}

// getSSID updates w.SSID value using the first listed wireless interface.
func (w *WifiData) getSSID() error {
	client, err := wifi.New()
	if err != nil {
		return fmt.Errorf("error creating new wifi client: %w", err)
	}

	interfaces, err := client.Interfaces()
	if err != nil {
		return fmt.Errorf("error listing network interfaces: %w", err)
	}

	var usableInterfaces []*wifi.Interface
	var primaryWifiInterface *wifi.Interface
	for _, f := range interfaces {
		if f.Name != "" {
			usableInterfaces = append(usableInterfaces, f)
		}
	}

	if len(usableInterfaces) == 0 {
		return errors.New("no usable wifi interfaces")
	}

	primaryWifiInterface = usableInterfaces[0]

	bss, err := client.BSS(primaryWifiInterface)
	if err != nil {
		return fmt.Errorf("error getting BSS: %w", err)
	}

	w.SSID = bss.SSID

	return nil
}
