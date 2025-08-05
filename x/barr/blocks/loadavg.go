package blocks

import (
	"fmt"
	"log"
	"syscall"
)

// See https://github.com/capnm/sysinfo/blob/master/sysinfo.go
const scale = 65536.0

// LoadAvg represents the system load average (1, 5, and 15 minutes).
type LoadAvg struct {
	load1  float64
	load5  float64
	load15 float64
}

// String implements the fmt.Stringer interface.
func (la *LoadAvg) String() string {
	si := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(si)
	if err != nil {
		log.Println("couldn't get sysinfo:", err)
		return ""
	}

	la.load1 = float64(si.Loads[0]) / scale
	la.load5 = float64(si.Loads[1]) / scale
	la.load15 = float64(si.Loads[2]) / scale
	return fmt.Sprintf("%.2f %.2f %.2f", la.load1, la.load5, la.load15)
}
