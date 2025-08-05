package blocks

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/distatus/battery"
)

// Battery represents battery information.
type Battery struct{}

// String returns battery info in a textual format.
func (b *Battery) String() string {
	batteries, err := battery.GetAll()
	if err != nil {
		return ""
	}

	if len(batteries) == 0 {
		return ""
	}

	var capsCur, capsFull float64
	for _, bat := range batteries {
		capsCur += bat.Current
		capsFull += bat.Full
	}
	pctBatRemaining := capsCur / capsFull * 100

	if c, err := Charging(); c && err == nil {
		return fmt.Sprintf("⚡AC.%s%%", strconv.FormatFloat(pctBatRemaining, 'f', 0, 64))
	}

	return fmt.Sprintf("🔋%s%%", strconv.FormatFloat(pctBatRemaining, 'f', 0, 64))
}

// Charging returns true if power supply AC is online.
func Charging() (bool, error) {
	online, err := ioutil.ReadFile("/sys/class/power_supply/AC/online")
	if err != nil {
		return false, err
	}
	charging, err := strconv.ParseBool(string(online[0]))
	return charging, err
}
