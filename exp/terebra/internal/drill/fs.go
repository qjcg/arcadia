package drill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
)

// drillFs drills into filesystem metadata.
func drillFs(args []string, stdout, stderr io.Writer) int {
	var recursive bool
	var followSymlinks bool
	var paths []string

	for i := range args {
		arg := args[i]
		switch {
		case arg == "-r" || arg == "--recursive":
			recursive = true
		case arg == "-l" || arg == "--follow":
			followSymlinks = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "drill fs: unknown flag: %s\n", arg)
			return 1
		default:
			paths = append(paths, arg)
		}
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	ctx := cueutil.NewContext()
	for _, path := range paths {
		if err := drillFsPath(ctx, path, recursive, followSymlinks, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "drill fs: %v\n", err)
			return 1
		}
	}
	return 0
}

func drillFsPath(ctx *cue.Context, path string, recursive, followSymlinks bool, stdout, stderr io.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	entry := fileInfoToCUE(ctx, path, info)
	str, err := cueutil.FormatValue(entry)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, str)

	if recursive && info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, e := range entries {
			childPath := filepath.Join(path, e.Name())
			if err := drillFsPath(ctx, childPath, recursive, followSymlinks, stdout, stderr); err != nil {
				fmt.Fprintf(stderr, "drill fs: %v\n", err)
			}
		}
	}
	return nil
}

func fileInfoToCUE(ctx *cue.Context, path string, info os.FileInfo) cue.Value {
	data := map[string]any{
		"path":    path,
		"name":    info.Name(),
		"size":    info.Size(),
		"mode":    info.Mode().String(),
		"isDir":   info.IsDir(),
		"modTime": info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
	}

	return ctx.Encode(data)
}
