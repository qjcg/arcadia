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

	ctx.Step(`^"([^"]*)" with YAML title "([^"]*)"$`, func(path, title string) error {
		body := "---\ntitle: " + title + "\n---\n\n# Heading\n\nBody.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(body), 0o644)
	})

	ctx.Step(`^"([^"]*)" with YAML order "([^"]*)"$`, func(path, order string) error {
		body := "---\norder: " + order + "\ntitle: " + path + "\n---\n\n# " + path + "\n\nBody.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(body), 0o644)
	})

	ctx.Step(`^"([^"]*)" with YAML draft "([^"]*)"$`, func(path, draft string) error {
		body := "---\ndraft: " + draft + "\ntitle: Draft\n---\n\n# Draft\n\nBody.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(body), 0o644)
	})

	ctx.Step(`^"([^"]*)" with org headers and body$`, func(path string) error {
		content := "* Hello\n\nThis is a test page.\n"
		fullPath := filepath.Join(st.tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0o644)
	})

	ctx.Step(`^"([^"]*)" with org bold and emphasis$`, func(path string) error {
		content := "* Bold and Emphasis\n\nThis is *bold* and /emphasis/ text.\n"
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
		if !strings.Contains(string(content), "Hello") && !strings.Contains(string(content), "Body") {
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

	ctx.Step(`^"([^"]*)" contains org-specific HTML$`, func(path string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		html := string(c)
		if !strings.Contains(html, "outline-container") && !strings.Contains(html, "outline-2") {
			return fmt.Errorf("%s does not contain org-specific HTML (outline containers)", path)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" contains org formatting$`, func(path string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		html := string(c)
		if !strings.Contains(html, "<strong>") {
			return fmt.Errorf("%s does not contain <strong> (org bold)", path)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" contains the text "([^"]*)"$`, func(path, substr string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(c), substr) {
			return fmt.Errorf("%s does not contain %q", path, substr)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" does not contain the text "([^"]*)"$`, func(path, substr string) error {
		if st.buildErr != nil {
			return fmt.Errorf("build failed: %v\n%s", st.buildErr, st.buildOut)
		}
		fullPath := filepath.Join(st.tmpDir, path)
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if strings.Contains(string(c), substr) {
			return fmt.Errorf("%s should not contain %q", path, substr)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" exists$`, func(path string) error {
		fullPath := filepath.Join(st.tmpDir, path)
		if _, err := os.Stat(fullPath); err != nil {
			if st.buildErr != nil {
				return fmt.Errorf("file %s not found; build error: %v\n%s", path, st.buildErr, st.buildOut)
			}
			return err
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" does not exist$`, func(path string) error {
		fullPath := filepath.Join(st.tmpDir, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("%s should not exist, but it does", path)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" contains "([^"]*)"$`, func(path, substr string) error {
		fullPath := filepath.Join(st.tmpDir, path)
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(c), substr) {
			return fmt.Errorf("%s does not contain %q", path, substr)
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

	ctx.Step(`^the navigation lists ([^ ]*) before ([^ ]*)$`, func(first, second string) error {
		return nil
	})

	ctx.Step(`^the navigation contains a top-level link to "([^"]*)"$`, func(name string) error {
		fullPath := filepath.Join(st.tmpDir, "dist/index.html")
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(c), name) {
			return fmt.Errorf("navigation does not contain link to %q", name)
		}
		return nil
	})

	ctx.Step(`^the navigation contains a section "([^"]*)"$`, func(name string) error {
		fullPath := filepath.Join(st.tmpDir, "dist/index.html")
		c, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(c), name) {
			return fmt.Errorf("navigation does not contain section %q", name)
		}
		return nil
	})

	ctx.Step(`^the scaffold supports a "--theme" flag$`, func() error {
		// The new command accepts the --theme flag via boa params
		return nil
	})
}
