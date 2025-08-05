package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/signintech/gopdf"
)

// PageSizeLetterLandscape provides landscape pages.
var PageSizeLetterLandscape = &gopdf.Rect{
	W: 792,
	H: 612,
}

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
func (d *DiplomaSet) Dump(w io.Writer) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", data)
}

// Load reads config from JSON file, populating a DiplomaSet.
func (d *DiplomaSet) Load(configFile string) {
	return
}

// ToPDF renders a DiplomaSet to PDF files.
// FIXME: Make this DRY by writing a utility function
func (d *DiplomaSet) ToPDF(fontFamily string, fontData []byte) {
	// Create OutputDir for PDFs
	os.MkdirAll(d.OutputDir, 0o700)

	for _, s := range d.Recipients {
		pdf := gopdf.GoPdf{}

		pdf.Start(gopdf.Config{PageSize: *PageSizeLetterLandscape})
		pdf.AddPage()

		err := pdf.AddTTFFontData(fontFamily, fontData)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: confirm this needs to be a .jpg, not .png?
		pdf.Image(d.Image, d.Overlay["Image"][0], d.Overlay["Image"][1], nil)

		// Recipient
		err = pdf.SetFont(fontFamily, "", 26)
		if err != nil {
			log.Fatal(err)
		}
		pdf.SetX(d.Overlay["Recipient"][0])
		pdf.SetY(d.Overlay["Recipient"][1])
		pdf.Cell(nil, s)

		// Course
		err = pdf.SetFont(fontFamily, "", 13)
		if err != nil {
			log.Fatal(err)
		}
		pdf.SetX(d.Overlay["Course"][0])
		pdf.SetY(d.Overlay["Course"][1])
		pdf.Cell(nil, d.Course)

		// Period
		err = pdf.SetFont(fontFamily, "", 11)
		if err != nil {
			log.Fatal(err)
		}
		pdf.SetX(d.Overlay["Period"][0])
		pdf.SetY(d.Overlay["Period"][1])
		pdf.Cell(nil, d.Period)

		// Instructor
		pdf.SetX(d.Overlay["Instructor"][0])
		pdf.SetY(d.Overlay["Instructor"][1])
		pdf.Cell(nil, d.Instructor)

		// FIXME: do this via a separate function for greater testability?
		pdfPath := filepath.Join(d.OutputDir, slug.Make(s)+".pdf")
		pdf.WritePdf(pdfPath)
	}
}
