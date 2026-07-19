package builtins

import (
	"fmt"
	"io"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
)

// CueHandler returns a handler for the "cue" builtin subcommand.
func CueHandler() Handler {
	return func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		if len(args) == 0 {
			fmt.Fprintln(stderr, "cue: expected subcommand (eval, vet, export, def, fmt, trim)")
			return 1
		}

		sub := args[0]
		subArgs := args[1:]

		switch sub {
		case "eval":
			return cueEval(subArgs, stdin, stdout, stderr)
		case "vet":
			return cueVet(subArgs, stdin, stdout, stderr)
		case "export":
			return cueExport(subArgs, stdin, stdout, stderr)
		case "def":
			return cueDef(subArgs, stdin, stdout, stderr)
		case "fmt":
			return cueFmt(subArgs, stdin, stdout, stderr)
		case "trim":
			return cueTrim(subArgs, stdin, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "cue: unknown subcommand %q (expected eval, vet, export, def, fmt, trim)\n", sub)
			return 1
		}
	}
}

// cueEval evaluates and prints a CUE file.
func cueEval(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx := cuecontext.New()
	var v cue.Value

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "cue eval: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "cue eval: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	}

	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "cue eval: %v\n", err)
		return 1
	}

	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "cue eval: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}

// cueVet validates data against a schema.
func cueVet(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "cue vet: expected <data> <schema>")
		return 1
	}

	ctx := cuecontext.New()

	dataPath := args[0]
	schemaPath := args[1]

	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		fmt.Fprintf(stderr, "cue vet: %v\n", err)
		return 1
	}

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "cue vet: %v\n", err)
		return 1
	}

	data := cueutil.CompileBytes(ctx, dataBytes)
	if err := cueutil.Err(data); err != nil {
		fmt.Fprintf(stderr, "cue vet: data: %v\n", err)
		return 1
	}

	schema := cueutil.CompileBytes(ctx, schemaBytes)
	if err := cueutil.Err(schema); err != nil {
		fmt.Fprintf(stderr, "cue vet: schema: %v\n", err)
		return 1
	}

	unified := cueutil.Unify(data, schema)
	if err := cueutil.Err(unified); err != nil {
		fmt.Fprintf(stderr, "cue vet: %v\n", err)
		return 1
	}

	if err := cueutil.Validate(unified); err != nil {
		fmt.Fprintf(stderr, "cue vet: validation failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "ok")
	return 0
}

// cueExport evaluates a CUE file and exports as JSON.
func cueExport(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx := cuecontext.New()
	var v cue.Value

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "cue export: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "cue export: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	}

	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "cue export: %v\n", err)
		return 1
	}

	jsonBytes, err := cueutil.ToJSON(v)
	if err != nil {
		fmt.Fprintf(stderr, "cue export: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(jsonBytes))
	return 0
}

// cueDef prints the consolidated definition of a CUE file.
func cueDef(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx := cuecontext.New()
	var v cue.Value

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "cue def: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "cue def: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	}

	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "cue def: %v\n", err)
		return 1
	}

	str, err := cueutil.FormatValueRaw(v)
	if err != nil {
		fmt.Fprintf(stderr, "cue def: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}

// cueFmt formats a CUE file.
func cueFmt(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx := cuecontext.New()
	var v cue.Value
	var source string

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "cue fmt: %v\n", err)
			return 1
		}
		source = string(data)
		v = cueutil.CompileBytes(ctx, data)
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "cue fmt: %v\n", err)
			return 1
		}
		source = string(data)
		v = cueutil.CompileBytes(ctx, data)
	}

	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "cue fmt: %v\n", err)
		return 1
	}

	// Format the CUE value
	formatted, err := cueutil.FormatValue(v)
	if err != nil {
		// Fallback to original
		fmt.Fprint(stdout, source)
		return 0
	}

	// Only write if different from original
	normalized := strings.TrimRight(formatted, "\n")
	if normalized != strings.TrimRight(source, "\n") {
		fmt.Fprintln(stdout, normalized)
	} else {
		fmt.Fprint(stdout, source)
	}
	return 0
}

// cueTrim removes redundant values from a CUE file.
func cueTrim(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// For now, cue trim just reformats the CUE output
	// Full trim support requires schema comparison
	ctx := cuecontext.New()
	var v cue.Value

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "cue trim: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	} else {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "cue trim: %v\n", err)
			return 1
		}
		v = cueutil.CompileBytes(ctx, data)
	}

	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "cue trim: %v\n", err)
		return 1
	}

	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "cue trim: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}
