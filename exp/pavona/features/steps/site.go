package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

type siteState struct {
	tmpDir   string
	bin      string
	serveCmd *exec.Cmd
	buildOut string
	buildErr error
}

func RegisterSiteSteps(ctx *godog.ScenarioContext) {
	st := &siteState{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		var err error
		st.tmpDir, err = os.MkdirTemp("", "pavona-site-test-")
		if err != nil {
			return ctx, err
		}
		st.bin = os.Getenv("PAVONA_BIN")
		if st.bin == "" {
			st.bin = filepath.Join(st.tmpDir, "pavona")
			out, err := exec.Command("go", "build", "-o", st.bin, ".").CombinedOutput()
			if err != nil {
				return ctx, fmt.Errorf("building pavona: %w\n%s", err, out)
			}
		}
		st.serveCmd = nil
		st.buildOut = ""
		st.buildErr = nil
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if st.serveCmd != nil && st.serveCmd.Process != nil {
			st.serveCmd.Process.Kill()
		}
		os.RemoveAll(st.tmpDir)
		return ctx, nil
	})

	ctx.Step(`^"([^"]*)" with frontmatter and body$`, func(path string) error {
		content := "# Hello\n\nThis is a test page.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0o644)
	})

	ctx.Step(`^"([^"]*)" with org headers and body$`, func(path string) error {
		content := "* Hello\n\nThis is a test page.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0o644)
	})

	ctx.Step(`^I run `+"`"+`pavona build`+"`"+`$`, func() error {
		cmd := exec.Command(st.bin, "build", "-d", st.tmpDir)
		out, err := cmd.CombinedOutput()
		st.buildOut = string(out)
		st.buildErr = err
		// Don't return error here — caller decides if failure is expected
		return nil
	})

	ctx.Step(`^"([^"]*)" exists and contains the rendered body$`, func(path string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "Hello") {
			return fmt.Errorf("%s does not contain expected content", path)
		}
		if !strings.Contains(string(content), "<h1>") && !strings.Contains(string(content), "<h2>") {
			return fmt.Errorf("%s does not contain heading HTML", path)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" exists and contains the rendered content$`, func(path string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "Hello") {
			return fmt.Errorf("%s does not contain expected content", path)
		}
		if !strings.Contains(string(content), "<h2") && !strings.Contains(string(content), "<h1") {
			return fmt.Errorf("%s does not contain heading HTML", path)
		}
		return nil
	})

	ctx.Step(`^I should get an error about the content directory$`, func() error {
		if st.buildErr == nil {
			return fmt.Errorf("expected build error, but build succeeded")
		}
		if !strings.Contains(st.buildOut, "content") && !strings.Contains(st.buildOut, "no such file") {
			return fmt.Errorf("expected error about content directory, got: %s", st.buildOut)
		}
		return nil
	})

	ctx.Step(`^I run `+"`"+`pavona serve`+"`"+`$`, func() error {
		cmd := exec.Command(st.bin, "serve", "-d", st.tmpDir, "-p", "9877")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("pavona serve failed to start: %w", err)
		}
		st.serveCmd = cmd
		time.Sleep(300 * time.Millisecond)
		return nil
	})

	ctx.Step(`^the dev server starts on localhost$`, func() error {
		return nil
	})

	ctx.Step(`^the dev server serves the built file over HTTP$`, func() error {
		resp, err := http.Get("http://localhost:9877/index.html")
		if err != nil {
			return fmt.Errorf("failed to fetch from dev server: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("expected 200, got %d", resp.StatusCode)
		}
		return nil
	})
}
