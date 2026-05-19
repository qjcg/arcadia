package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

//go:embed fonts
var fonts embed.FS

var (
	stdout = os.Stdout
	stderr = os.Stderr
)

type GenerateParams struct {
	Course     string   `descr:"Course name"`
	Instructor string   `descr:"Instructor name" default:"Rory Q. Teachalot"`
	Period     string   `descr:"Training period/dates"`
	Recipients []string `descr:"Comma-separated recipients" default:"Joe Learnery"`
	OutputDir  string   `descr:"Output directory" default:"diplomas"`
	DryRun     bool     `descr:"Preview without generating PDFs"`
	Json       bool     `descr:"Output results as JSON"`
}

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "diplomat",
		Short:   "Generate PDF diplomas",
		Long:    "Generate PDF certificates of completion for training participants.",
		Version: getVersion(),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")
	root.PersistentFlags().BoolP("verbose", "v", false, "Show detailed progress")
	root.PersistentFlags().Bool("no-color", false, "Disable color output")

	root.AddCommand(generateCmd())

	return root
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func generateCmd() *cobra.Command {
	return boa.CmdT[GenerateParams]{
		Use:   "generate",
		Short: "Generate diplomas",
		RunFuncE: func(p *GenerateParams, cmd *cobra.Command, _ []string) error {
			return runGenerate(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runGenerate(p *GenerateParams, cmd *cobra.Command) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	verbose, _ := cmd.Flags().GetBool("verbose")

	recipients := p.Recipients
	if len(p.Recipients) == 1 && strings.Contains(p.Recipients[0], ",") {
		recipients = strings.Split(p.Recipients[0], ",")
	}

	session := &Session{
		Course:     p.Course,
		Period:     p.Period,
		Instructor: p.Instructor,
		Recipients: recipients,
	}

	d := &DiplomaSet{
		Session:   *session,
		Template:  Template{},
		OutputDir: filepath.Join(p.OutputDir, slug.Make(p.Course)),
	}

	if p.DryRun {
		if !quiet {
			fmt.Fprintf(stdout, "Would generate diplomas for %d recipients in %s/\n",
				len(session.Recipients), d.OutputDir)
		}
		return nil
	}

	fontData, err := fonts.ReadFile("fonts/DroidSans.ttf")
	if err != nil {
		return fmt.Errorf("loading font: %w", err)
	}

	if err := d.ToPDF("DroidSans", fontData); err != nil {
		return fmt.Errorf("generating diplomas: %w", err)
	}

	if p.Json {
		if err := d.Dump(stdout); err != nil {
			return fmt.Errorf("outputting JSON: %w", err)
		}
	}

	if verbose && !quiet {
		fmt.Fprintf(stdout, "Generated %d diplomas in %s/\n",
			len(session.Recipients), d.OutputDir)
	}

	return nil
}
