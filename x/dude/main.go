// Print man page table of contents or named sections.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

func main() {
	var out []byte
	var err error

	// subject: e.g. bash, find
	// section: e.g. 1, 5, 9
	// heading: e.g. EXAMPLES, BUGS
	var subject, section, heading string

	// Detect the arguments (excluding the command itself) and call the man
	// command accordingly.
	switch len(os.Args) - 1 {
	case 1:
		subject = os.Args[1]
		out, err = exec.Command("man", subject).Output()
	case 2:

		// With two arguments, there are two possible syntaxes:
		//   1. section + subject (ex: "5 passwd")
		//   2. subject + heading (ex: "passwd FILES")

		// Using the section + subject syntax.
		if matched, _ := regexp.MatchString(`\d+`, os.Args[1]); matched {
			section = os.Args[1]
			subject = os.Args[2]
			out, err = exec.Command("man", section, subject).Output()
			break
		}

		// Using the subject + heading syntax.
		subject = os.Args[1]
		heading = os.Args[2]
		out, err = exec.Command("man", subject).Output()
	case 3:
		section = os.Args[1]
		subject = os.Args[2]
		heading = os.Args[3]
		out, err = exec.Command("man", section, subject).Output()
	default:
		log.Fatal("USAGE MESSAGE")
	}
	if err != nil {
		log.Fatalf("Error calling \"man\" command: %s, out: %s\n", err, out)
	}

	// Regular expressions defining what is considered a heading line in
	// "man" output.
	Heading := regexp.MustCompile(`^( {3})?[A-Z]`)
	NotHeading := regexp.MustCompile(` {4}`)

	// Whether we are printing the contents of a heading (ex: EXAMPLES).
	var printHeadingContents bool

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		b := scanner.Bytes()

		// Line is a heading.
		if Heading.Match(b) && !NotHeading.Match(b) {

			// We set printSection to true the FIRST time a heading is
			// seen, so this conditional will be true ONLY when we
			// hit a SECOND section, and thus we want to exit.
			if printHeadingContents {
				os.Exit(0)
			}

			// When a heading argument has been passed on the
			// command line, set boolean indicating we want to
			// print the heading contents. The heading match is
			// case insensitive.
			if heading != "" {
				if m, _ := regexp.MatchString(`(?i)`+string(b), heading); m {
					printHeadingContents = true
				} else {
					continue
				}
			}
		} else {
			// When the line is not a heading and we're not
			// printing heading contents, move on to next line.
			if !printHeadingContents {
				continue
			}
		}

		// Print the current line.
		fmt.Printf("%s\n", b)
	}
}
