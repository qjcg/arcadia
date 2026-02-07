package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/arcadia/x/slidesdeck/assets"
	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/parser"
	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/ui"
	"github.com/spf13/cobra"
)

var (
	outputPath   string
	defaultTheme string
	separator    string
)

var rootCmd = &cobra.Command{
	Use:   "slidesdeck [input file]",
	Short: "Transform Markdown or Org-mode notes into interactive HTML slideshows",
	Long: `Slidesdeck is a single-purpose CLI tool that transforms your Markdown
or Emacs Org-mode notes into professional, self-contained HTML slideshows.

Slides are separated by first-level headings (# in Markdown, * in Org-mode)
by default, or by explicit horizontal rules (--- or -----).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Determine input format
		ext := strings.ToLower(filepath.Ext(inputFile))
		if ext != ".md" && ext != ".org" {
			return fmt.Errorf("unsupported file extension: %s (must be .md or .org)", ext)
		}

		// Determine final output path
		finalOutput := outputPath
		if finalOutput == "" {
			// Default: same directory as input, same name, .html extension
			finalOutput = strings.TrimSuffix(inputFile, ext) + ".html"
		}

		fmt.Printf("Converting %s to %s...\n", inputFile, finalOutput)

		f, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		opts := parser.Options{
			Separator: separator,
		}

		var deck *parser.Deck
		if ext == ".md" {
			deck, err = parser.ParseMarkdown(f, opts)
		} else {
			deck, err = parser.ParseOrg(f, opts)
		}
		if err != nil {
			return err
		}

		fmt.Printf("Generated %d slides\n", len(deck.Slides))

		// Load assets
		cssFile, err := assets.Dist.Open("dist/main.css")
		if err != nil {
			return err
		}
		defer cssFile.Close()
		cssData, _ := io.ReadAll(cssFile)

		jsFile, err := assets.Dist.Open("dist/main.js")
		if err != nil {
			return err
		}
		defer jsFile.Close()
		jsData, _ := io.ReadAll(jsFile)

		// Create output file
		out, err := os.Create(finalOutput)
		if err != nil {
			return err
		}
		defer out.Close()

		return ui.Layout(deck, defaultTheme, string(cssData), string(jsData)).Render(context.Background(), out)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: same as input with .html extension)")
	rootCmd.Flags().StringVarP(&defaultTheme, "theme", "t", "dark", "Default daisyUI theme")
	rootCmd.Flags().StringVarP(&separator, "separator", "s", "", "Custom slide separator (overrides headings)")
}
