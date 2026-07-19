package drill

import (
	"fmt"
	"io"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
)

// Handler is the drill command dispatcher.
func Handler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "drill: expected subcommand")
		fmt.Fprintln(stderr, "  drill cue <file>  Drill into CUE files")
		fmt.Fprintln(stderr, "  drill fs <path>   Drill into filesystem (planned)")
		fmt.Fprintln(stderr, "  drill proc <pid>  Drill into processes (planned)")
		fmt.Fprintln(stderr, "  drill net <host>  Drill into network (planned)")
		return 1
	}

	sub := args[0]
	subArgs := args[1:]

	switch sub {
	case "cue":
		return drillCue(subArgs, stdin, stdout, stderr)
	case "fs":
		return drillFs(subArgs, stdout, stderr)
	case "proc":
		return drillProc(subArgs, stdout, stderr)
	case "net":
		return drillNet(subArgs, stdout, stderr)
	case "help":
		fmt.Fprintln(stdout, "drill: drill into structured data")
		fmt.Fprintln(stdout, "  drill cue <file>  Evaluate, validate, and walk CUE values")
		fmt.Fprintln(stdout, "    -e 'path'      Extract a path from the value")
		fmt.Fprintln(stdout, "    -v             Validate against schema")
		fmt.Fprintln(stdout, "    --export json  Export as JSON")
		fmt.Fprintln(stdout, "  drill fs <path>  Drill into filesystem metadata")
		fmt.Fprintln(stdout, "    -r             Recursive")
		fmt.Fprintln(stdout, "    -l             Follow symlinks")
		fmt.Fprintln(stdout, "  drill proc <pid> Drill into process information")
		fmt.Fprintln(stdout, "  drill net <host> Drill into network connections")
		return 0
	default:
		fmt.Fprintf(stderr, "drill: unknown subcommand: %s\n", sub)
		return 1
	}
}

func drillCue(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Parse flags
	var extractPath string
	var validate bool
	var exportJSON bool
	var unifyFile string
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-e" || arg == "--extract":
			i++
			if i >= len(args) {
				fmt.Fprintln(stderr, "drill cue: missing path for -e")
				return 1
			}
			extractPath = args[i]
		case arg == "-v" || arg == "--validate":
			validate = true
		case arg == "--export" && i+1 < len(args) && args[i+1] == "json":
			exportJSON = true
			i++
		case arg == "--unify" && i+1 < len(args):
			i++
			unifyFile = args[i]
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "drill cue: unknown flag: %s\n", arg)
			return 1
		default:
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		// Read from stdin (the pipe)
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "drill cue: %v\n", err)
			return 1
		}
		src := strings.TrimSpace(string(data))
		if src == "" {
			return 0
		}
		return drillCueString(src, extractPath, validate, exportJSON, unifyFile, stdout, stderr)
	}

	// Read from files
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(stderr, "drill cue: %v\n", err)
			return 1
		}
		src := strings.TrimSpace(string(data))
		if src == "" {
			continue
		}
		code := drillCueString(src, extractPath, validate, exportJSON, unifyFile, stdout, stderr)
		if code != 0 {
			return code
		}
	}
	return 0
}

func drillCueString(src, extractPath string, validate, exportJSON bool, unifyFile string, stdout, stderr io.Writer) int {
	ctx := cueutil.NewContext()
	v := cueutil.CompileString(ctx, src)
	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "drill cue: %v\n", err)
		return 1
	}

	// Unify with another file if requested
	if unifyFile != "" {
		unifyData, err := os.ReadFile(unifyFile)
		if err != nil {
			fmt.Fprintf(stderr, "drill cue: --unify: %v\n", err)
			return 1
		}
		unifyV := cueutil.CompileBytes(ctx, unifyData)
		if err := cueutil.Err(unifyV); err != nil {
			fmt.Fprintf(stderr, "drill cue: --unify: %v\n", err)
			return 1
		}
		v = cueutil.Unify(v, unifyV)
		if err := cueutil.Err(v); err != nil {
			fmt.Fprintf(stderr, "drill cue: unify: %v\n", err)
			return 1
		}
	}

	return drillCueValue(v, extractPath, validate, exportJSON, stdout, stderr)
}

func drillCueValue(v cue.Value, extractPath string, validate, exportJSON bool, stdout, stderr io.Writer) int {
	// Extract path if requested
	if extractPath != "" {
		v = cueutil.LookupPath(v, extractPath)
		if err := cueutil.Err(v); err != nil {
			fmt.Fprintf(stderr, "drill cue: path %q: %v\n", extractPath, err)
			return 1
		}
	}

	// Validate if requested
	if validate {
		if err := cueutil.Validate(v); err != nil {
			fmt.Fprintf(stderr, "drill cue: validation error: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "ok")
		return 0
	}

	// Export as JSON if requested
	if exportJSON {
		b, err := cueutil.ToJSON(v)
		if err != nil {
			fmt.Fprintf(stderr, "drill cue: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(b))
		return 0
	}

	// Default: format and print
	str, err := cueutil.FormatValue(v)
	if err != nil {
		str, err = cueutil.FormatValueRaw(v)
		if err != nil {
			fmt.Fprintf(stderr, "drill cue: %v\n", err)
			return 1
		}
	}
	fmt.Fprintln(stdout, str)
	return 0
}
