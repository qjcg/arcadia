package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// Letter landscape in points (792 x 612)
const (
	pageWidth  = 792.0
	pageHeight = 612.0
)

// Session represents a training session.
type Session struct {
	Course     string
	Period     string
	Instructor string
	Recipients []string
}

// Template contains an image file path along with a map of text overlay
// coordinates.
type Template struct {
	Image   string
	Overlay map[string][2]float64
}

// DiplomaSet contains an OutputDir for PDFs, and embedded Template and Session
// structs.
type DiplomaSet struct {
	Session
	Template
	OutputDir string
}

// Dump writes JSON config to an io.Writer.
func (d *DiplomaSet) Dump(w io.Writer) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintf(w, "%s\n", data)
	return nil
}

// Load reads config from JSON file, populating a DiplomaSet.
func (d *DiplomaSet) Load(configFile string) {
}

// ToPDF renders a DiplomaSet to PDF files.
func (d *DiplomaSet) ToPDF(fontFamily string, fontData []byte) error {
	if err := os.MkdirAll(d.OutputDir, 0o700); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, recipient := range d.Recipients {
		doc := template.New(
			template.WithPageSize(document.Size{Width: pageWidth, Height: pageHeight}),
			template.WithFont(fontFamily, fontData),
			template.WithDefaultFont(fontFamily, 12),
			template.WithMargins(document.Edges{
				Top:    document.Pt(0),
				Bottom: document.Pt(0),
				Left:   document.Pt(0),
				Right:  document.Pt(0),
			}),
		)

		page := doc.AddPage()

		// Background - cream/off-white (fills full page)
		page.Absolute(document.Pt(0), document.Pt(0), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(pageWidth)),
				template.WithBoxHeight(document.Pt(pageHeight)),
				template.WithBoxBackground(pdf.RGBHex(0xFFFEF5)),
			)
		}, template.AbsoluteOriginPage())

		// Navy outer border - using borders instead of background fill
		// Top: y=8, height=4 -> PDF y=600 (8pt from top: 612-600-4=8) ✓
		page.Absolute(document.Pt(8), document.Pt(8), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(784)),
				template.WithBoxHeight(document.Pt(4)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(4)),
					template.BorderColor(pdf.RGBHex(0x1A365D)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Bottom: code y=600 -> PDF y=8 (8pt from bottom: 612-8-4=8) ✓
		page.Absolute(document.Pt(8), document.Pt(600), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(784)),
				template.WithBoxHeight(document.Pt(4)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(4)),
					template.BorderColor(pdf.RGBHex(0x1A365D)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Left: x=8, width=4 (4pt thick bar at x=8 to x=12)
		page.Absolute(document.Pt(8), document.Pt(8), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(4)),
				template.WithBoxHeight(document.Pt(596)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(4)),
					template.BorderColor(pdf.RGBHex(0x1A365D)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Right: x=788, width=4 (4pt wide bar at x=788 to x=792)
		page.Absolute(document.Pt(788), document.Pt(8), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(4)),
				template.WithBoxHeight(document.Pt(596)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(4)),
					template.BorderColor(pdf.RGBHex(0x1A365D)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Gold inner border - using borders
		// Top: y=596.5
		page.Absolute(document.Pt(14), document.Pt(596.5), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(778)),
				template.WithBoxHeight(document.Pt(1.5)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1.5)),
					template.BorderColor(pdf.RGBHex(0xC9A227)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Bottom: y=14
		page.Absolute(document.Pt(14), document.Pt(14), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(778)),
				template.WithBoxHeight(document.Pt(1.5)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1.5)),
					template.BorderColor(pdf.RGBHex(0xC9A227)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Left: x=14, width=1.5, height=582
		page.Absolute(document.Pt(14), document.Pt(14), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(1.5)),
				template.WithBoxHeight(document.Pt(582)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1.5)),
					template.BorderColor(pdf.RGBHex(0xC9A227)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// Right: x=776.5, width=1.5, height=582
		page.Absolute(document.Pt(776.5), document.Pt(14), func(c *template.ColBuilder) {
			c.Box(
				func(c *template.ColBuilder) {},
				template.WithBoxWidth(document.Pt(1.5)),
				template.WithBoxHeight(document.Pt(582)),
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1.5)),
					template.BorderColor(pdf.RGBHex(0xC9A227)),
				)),
			)
		}, template.AbsoluteOriginPage())

		// "CERTIFICATE OF COMPLETION" header - centered
		page.Absolute(document.Pt(246), document.Pt(40), func(c *template.ColBuilder) {
			c.Text(
				"CERTIFICATE OF COMPLETION",
				template.FontSize(24),
				template.Bold(),
				template.TextColor(pdf.RGBHex(0x1A365D)),
			)
		}, template.AbsoluteWidth(document.Pt(300)))

		// Gold decorative line under header - centered, 500pt wide
		page.Absolute(document.Pt(146), document.Pt(72), func(c *template.ColBuilder) {
			c.Line(
				template.LineColor(pdf.RGBHex(0xC9A227)),
				template.LineThickness(document.Pt(2)),
			)
		}, template.AbsoluteWidth(document.Pt(500)))

		// "This is to certify that" - centered
		page.Absolute(document.Pt(296), document.Pt(95), func(c *template.ColBuilder) {
			c.Text(
				"This is to certify that",
				template.FontSize(12),
				template.TextColor(pdf.RGBHex(0x4A5568)),
			)
		}, template.AbsoluteWidth(document.Pt(200)))

		// Recipient name - centered
		page.Absolute(document.Pt(156), document.Pt(125), func(c *template.ColBuilder) {
			c.Text(
				recipient,
				template.FontSize(32),
				template.Bold(),
				template.TextColor(pdf.RGBHex(0x1A365D)),
			)
		}, template.AbsoluteWidth(document.Pt(480)))

		// Gold underline for name - centered, 400pt wide
		page.Absolute(document.Pt(196), document.Pt(162), func(c *template.ColBuilder) {
			c.Line(
				template.LineColor(pdf.RGBHex(0xC9A227)),
				template.LineThickness(document.Pt(2)),
			)
		}, template.AbsoluteWidth(document.Pt(400)))

		// "has successfully completed" - centered
		page.Absolute(document.Pt(246), document.Pt(180), func(c *template.ColBuilder) {
			c.Text(
				"has successfully completed the course",
				template.FontSize(12),
				template.TextColor(pdf.RGBHex(0x4A5568)),
			)
		}, template.AbsoluteWidth(document.Pt(300)))

		// Course name - centered
		page.Absolute(document.Pt(246), document.Pt(210), func(c *template.ColBuilder) {
			c.Text(
				d.Course,
				template.FontSize(22),
				template.Bold(),
				template.TextColor(pdf.RGBHex(0x2D3748)),
			)
		}, template.AbsoluteWidth(document.Pt(300)))

		// Period/hours - centered
		page.Absolute(document.Pt(316), document.Pt(242), func(c *template.ColBuilder) {
			c.Text(
				d.Period,
				template.FontSize(10),
				template.Italic(),
				template.TextColor(pdf.RGBHex(0x718096)),
			)
		}, template.AbsoluteWidth(document.Pt(160)))

		// Signature section - Instructor label
		page.Absolute(document.Pt(80), document.Pt(340), func(c *template.ColBuilder) {
			c.Text(
				"Instructor",
				template.FontSize(9),
				template.TextColor(pdf.RGBHex(0x718096)),
			)
		}, template.AbsoluteWidth(document.Pt(100)))

		// Instructor signature line
		page.Absolute(document.Pt(60), document.Pt(355), func(c *template.ColBuilder) {
			c.Line(
				template.LineColor(pdf.RGBHex(0x1A365D)),
				template.LineThickness(document.Pt(1)),
			)
		}, template.AbsoluteWidth(document.Pt(140)))

		// Instructor name
		page.Absolute(document.Pt(80), document.Pt(362), func(c *template.ColBuilder) {
			c.Text(
				d.Instructor,
				template.FontSize(12),
				template.Bold(),
				template.TextColor(pdf.RGBHex(0x2D3748)),
			)
		}, template.AbsoluteWidth(document.Pt(140)))

		// Date label - right side
		page.Absolute(document.Pt(590), document.Pt(340), func(c *template.ColBuilder) {
			c.Text(
				"Date",
				template.FontSize(9),
				template.TextColor(pdf.RGBHex(0x718096)),
			)
		}, template.AbsoluteWidth(document.Pt(80)))

		// Date signature line - right side
		page.Absolute(document.Pt(570), document.Pt(355), func(c *template.ColBuilder) {
			c.Line(
				template.LineColor(pdf.RGBHex(0x1A365D)),
				template.LineThickness(document.Pt(1)),
			)
		}, template.AbsoluteWidth(document.Pt(120)))

		// Date placeholder - right side
		page.Absolute(document.Pt(570), document.Pt(362), func(c *template.ColBuilder) {
			c.Text(
				"____________________",
				template.FontSize(10),
				template.TextColor(pdf.RGBHex(0x2D3748)),
			)
		}, template.AbsoluteWidth(document.Pt(120)))

		pdfPath := filepath.Join(d.OutputDir, slug.Make(recipient)+".pdf")
		f, err := os.Create(pdfPath)
		if err != nil {
			return fmt.Errorf("creating PDF file: %w", err)
		}
		data, err := doc.Generate()
		if err != nil {
			f.Close()
			return fmt.Errorf("generating PDF: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			return fmt.Errorf("writing PDF: %w", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("closing PDF: %w", err)
		}
	}
	return nil
}
