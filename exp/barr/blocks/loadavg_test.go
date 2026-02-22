package blocks

import (
	"regexp"
	"testing"
)

func TestLoadAvg(t *testing.T) {
	var la LoadAvg
	curLoadAvg := la.String()
	if m, _ := regexp.MatchString("([0-9]+.[0-9]{2} ?){3}", curLoadAvg); !m {
		t.Fatalf("actual output does not match regexp: %s", curLoadAvg)
	}
	t.Attr("curLoadAvg", curLoadAvg)
}
