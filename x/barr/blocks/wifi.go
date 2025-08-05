package blocks

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

type WifiData struct {
	ESSID string
}

// String implements the fmt.Stringer interface.
func (w *WifiData) String() string {
	err := w.getESSID()
	if err != nil || w.ESSID == "" {
		return ""
	}

	return w.ESSID
}

// getESSID updates w.ESSID value based on the output of the "iw dev" command.
//
// NOTE: "iw" help output says: "Do NOT screenscrape this tool, we don't
// consider its output stable." Until a better solution comes along, we'll
// screenscrape it anyway!
func (w *WifiData) getESSID() error {
	out, err := exec.Command("iw", "dev").Output()
	if e, ok := err.(*exec.ExitError); ok {
		return e
	}
	if err != nil {
		return err
	}

	r := bytes.NewReader(out)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ssid") {
			w.ESSID = strings.Join(strings.Fields(line)[1:], " ")
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}
