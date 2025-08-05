package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
)

//go:embed fonts
var fonts embed.FS

func main() {
	course := flag.String("c", "", "course name")
	outDir := flag.String("d", "diplomas", "output directory")
	period := flag.String("p", "", "training period (dates)")
	recipients := flag.String("r", "Joe Learnery", "list of recipients (comma-separated)")
	instructor := flag.String("i", "Rory Q. Teachalot", "instructor's name")

	// TODO: add template flag handling logic
	// template := flag.String("t", "default", "template")
	flag.Parse()

	// TODO: use cli template option if provided
	template := DefaultTemplate

	session := &Session{
		Course:     *course,
		Period:     *period,
		Instructor: *instructor,
		Recipients: strings.Split(*recipients, ","),
	}

	d := &DiplomaSet{
		Session:   *session,
		Template:  *template,
		OutputDir: filepath.Join(*outDir, slug.Make(*course)),
	}
	fontData, err := fonts.ReadFile(("fonts/DroidSans.ttf"))
	if err != nil {
		log.Fatal(err)
	}

	d.ToPDF("DroidSans", fontData)

	// create JSON config file
	confPath := filepath.Join(d.OutputDir, "diplomas.json")
	conf, err := os.Create(confPath)
	defer conf.Close()
	if err != nil {
		log.Fatal(err)
	}
	d.Dump(conf)
}
