package drill

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestHandlerNoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := Handler(nil, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "expected subcommand") {
		t.Fatalf("expected usage error, got %q", errb.String())
	}
}

func TestHandlerUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	code := Handler([]string{"bogus"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand error, got %q", errb.String())
	}
}

func TestHandlerHelp(t *testing.T) {
	var out, errb bytes.Buffer
	code := Handler([]string{"help"}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "drill: drill into structured data") {
		t.Fatalf("expected help text, got %q", out.String())
	}
}

func TestDrillCueFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	src := "a: 1\nb: \"two\"\n"
	code := drillCue(nil, strings.NewReader(src), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "a:") {
		t.Fatalf("expected formatted CUE output, got %q", out.String())
	}
}

func TestDrillCueFromStdinEmpty(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillCue(nil, strings.NewReader("   \n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0 for empty input, got %d", code)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output for empty input, got %q", out.String())
	}
}

func TestDrillCueInvalidSource(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillCue(nil, strings.NewReader("a: {"), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "drill cue:") {
		t.Fatalf("expected drill cue error, got %q", errb.String())
	}
}

func TestDrillCueFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.cue")
	if err := os.WriteFile(file, []byte("x: 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := drillCue([]string{file}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "x:") {
		t.Fatalf("expected formatted output, got %q", out.String())
	}
}

func TestDrillCueMissingFile(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillCue([]string{"/nonexistent/file.cue"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestDrillCueExtract(t *testing.T) {
	var out, errb bytes.Buffer
	src := "a: { b: 1 }\n"
	code := drillCue([]string{"-e", "a.b"}, strings.NewReader(src), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "1") {
		t.Fatalf("expected extracted value, got %q", out.String())
	}
}

func TestDrillCueExtractMissingPath(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillCue([]string{"-e"}, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "missing path") {
		t.Fatalf("expected missing path error, got %q", errb.String())
	}
}

func TestDrillCueValidate(t *testing.T) {
	var out, errb bytes.Buffer
	src := "a: 1\n"
	code := drillCue([]string{"-v"}, strings.NewReader(src), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected ok output, got %q", out.String())
	}
}

func TestDrillCueExportJSON(t *testing.T) {
	var out, errb bytes.Buffer
	src := "a: 1\n"
	code := drillCue([]string{"--export", "json"}, strings.NewReader(src), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), `"a"`) {
		t.Fatalf("expected JSON output, got %q", out.String())
	}
}

func TestDrillCueUnknownFlag(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillCue([]string{"--bogus"}, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown flag") {
		t.Fatalf("expected unknown flag error, got %q", errb.String())
	}
}

func TestDrillFs(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	code := drillFs([]string{dir}, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "isDir") {
		t.Fatalf("expected fs metadata output, got %q", out.String())
	}
}

func TestDrillFsDefaultPath(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillFs(nil, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected output for default path")
	}
}

func TestDrillFsUnknownFlag(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillFs([]string{"--bogus"}, &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestDrillFsMissingPath(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillFs([]string{"/nonexistent/path"}, &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestDrillFsRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := drillFs([]string{"-r", dir}, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "sub") {
		t.Fatalf("expected recursive output to include subdir, got %q", out.String())
	}
}

func TestDrillProcNoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillProc(nil, &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestDrillProcInvalidPID(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillProc([]string{"abc"}, &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "invalid PID") {
		t.Fatalf("expected invalid PID error, got %q", errb.String())
	}
}

func TestDrillProcNotFound(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillProc([]string{"99999999"}, &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestDrillNetLocal(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillNetLocal(&out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "hostname") {
		t.Fatalf("expected hostname in output, got %q", out.String())
	}
}

func TestDrillNetNoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillNet(nil, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestDrillNetInvalidHostPort(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillNetConnect("not-a-valid-host:port:extra", &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "invalid host:port") {
		t.Fatalf("expected invalid host:port error, got %q", errb.String())
	}
}

func TestDrillNetDNS(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillNetDNS("localhost", &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "host") {
		t.Fatalf("expected host in output, got %q", out.String())
	}
}

func TestDrillProcCurrentProcess(t *testing.T) {
	var out, errb bytes.Buffer
	code := drillProc([]string{strconv.Itoa(os.Getpid())}, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "name") {
		t.Fatalf("expected proc metadata, got %q", out.String())
	}
}
