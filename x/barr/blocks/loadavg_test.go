package blocks

import (
	"regexp"
	"testing"
)

func TestLoadAvg(t *testing.T) {
	var la LoadAvg
	curLa := la.String()
	if m, _ := regexp.MatchString("([0-9]+.[0-9]{2} ?){3}", curLa); !m {
		t.Fatalf("actual output does not match regexp: %s", curLa)
	}
	t.Logf("Current load average: %s\n", curLa)
}
