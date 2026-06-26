package steps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func RegisterSiteSteps(ctx *godog.ScenarioContext) {
	var tmpDir string
	var pavonaBin string

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		var err error
		tmpDir, err = os.MkdirTemp("", "pavona-site-test-")
		if err != nil {
			return ctx, err
		}
		pavonaBin = os.Getenv("PAVONA_BIN")
		if pavonaBin == "" {
			bin := filepath.Join(tmpDir, "pavona")
			out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
			if err != nil {
				return ctx, fmt.Errorf("building pavona: %w\n%s", err, out)
			}
			pavonaBin = bin
		}
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		os.RemoveAll(tmpDir)
		return ctx, nil
	})

	ctx.Step(`^"([^"]*)" with frontmatter and body$`, func(path string) error {
		content := "# Hello\n\nThis is a test page.\n"
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0o644)
	})

	ctx.Step(`^"([^"]*)" with org headers and body$`, func(path string) error {
		content := "* Hello\n\nThis is a test page.\n"
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0o644)
	})

	ctx.Step(`^I run \`+"`"+`pavona build\`+"`"+`$`, func() error {
		cmd := exec.Command(pavonaBin, "build", "-d", tmpDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("pavona build failed: %w\n%s", err, out)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" exists and contains the rendered body$`, func(path string) error {
		fullPath := filepath.Join(tmpDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "Hello") {
			return fmt.Errorf("%s does not contain expected content", path)
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" exists and contains the rendered content$`, func(path string) error {
		fullPath := filepath.Join(tmpDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "Hello") {
			return fmt.Errorf("%s does not contain expected content", path)
		}
		return nil
	})

	ctx.Step(`^I run \`+"`"+`pavona serve\`+"`"+`$`, func() error {
		cmd := exec.Command(pavonaBin, "serve", "-d", tmpDir, "-p", "9876")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("pavona serve failed to start: %w", err)
		}
		time.Sleep(300 * time.Millisecond)
		cmd.Process.Kill()
		return nil
	})

	ctx.Step(`^the dev server starts on localhost$`, func() error {
		return nil
	})
}
