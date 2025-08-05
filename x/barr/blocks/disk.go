package blocks

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DiskInfo represents hard disk info.
type Disk struct {
	Dir string
}

func (d *Disk) String() string {
	output, err := exec.Command("df", "-h", d.Dir).Output()
	if err != nil {
		return "df error"
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var i int
	for scanner.Scan() {
		// Skip header line.
		if i == 0 {
			i++
			continue
		}

		line := scanner.Text()
		avail := strings.Fields(line)[3]
		return fmt.Sprintf("d:%s", avail)
	}

	// Return the empty string if no usage lines scanned.
	return ""
}
