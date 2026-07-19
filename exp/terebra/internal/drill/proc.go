package drill

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
)

// drillProc drills into process information.
func drillProc(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "drill proc: expected PID")
		return 1
	}

	pidStr := args[0]
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Fprintf(stderr, "drill proc: invalid PID: %s\n", pidStr)
		return 1
	}

	procPath := fmt.Sprintf("/proc/%d", pid)

	// Check if process exists
	if _, err := os.Stat(procPath); err != nil {
		fmt.Fprintf(stderr, "drill proc: process %d not found\n", pid)
		return 1
	}

	ctx := cueutil.NewContext()
	procData := make(map[string]any)
	procData["pid"] = pid

	// Read /proc/pid/status
	if data, err := os.ReadFile(fmt.Sprintf("%s/status", procPath)); err == nil {
		lines := strings.SplitSeq(string(data), "\n")
		for line := range lines {
			if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				switch key {
				case "Name":
					procData["name"] = val
				case "State":
					procData["state"] = val
				case "Pid":
					procData["pid"], _ = strconv.Atoi(val)
				case "PPid":
					procData["ppid"], _ = strconv.Atoi(val)
				case "Uid":
					procData["uid"] = val
				case "Gid":
					procData["gid"] = val
				case "VmRSS":
					procData["memory"] = val
				case "Threads":
					procData["threads"], _ = strconv.Atoi(val)
				}
			}
		}
	}

	// Read /proc/pid/cmdline
	if data, err := os.ReadFile(fmt.Sprintf("%s/cmdline", procPath)); err == nil {
		cmdline := strings.ReplaceAll(string(data), "\x00", " ")
		procData["cmdline"] = strings.TrimSpace(cmdline)
	}

	// Read /proc/pid/cwd (symlink)
	if target, err := os.Readlink(fmt.Sprintf("%s/cwd", procPath)); err == nil {
		procData["cwd"] = target
	}

	// Read /proc/pid/exe (symlink)
	if target, err := os.Readlink(fmt.Sprintf("%s/exe", procPath)); err == nil {
		procData["exe"] = target
	}

	v := ctx.Encode(procData)
	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "drill proc: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}
