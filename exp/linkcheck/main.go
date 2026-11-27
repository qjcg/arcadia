// Command linkcheck validates links in tracked, first-party Markdown.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

var (
	markdown   = goldmark.New(goldmark.WithExtensions(extension.GFM))
	htmlAnchor = regexp.MustCompile(`(?i)\b(?:id|name)\s*=\s*["']([^"']+)["']`)
)

type document struct {
	anchors map[string]bool
	links   []string
}

func parse(source []byte) document {
	d := document{anchors: map[string]bool{}}
	headings := map[string]bool{}
	_ = ast.Walk(markdown.Parser().Parse(text.NewReader(source)), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Link:
			d.links = append(d.links, string(n.Destination))
		case *ast.Image:
			d.links = append(d.links, string(n.Destination))
		case *ast.AutoLink:
			if n.AutoLinkType != ast.AutoLinkEmail {
				d.links = append(d.links, string(n.URL(source)))
			}
		case *ast.Heading:
			slug := strings.Map(func(r rune) rune {
				switch {
				case r == ' ':
					return '-'
				case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r):
					return unicode.ToLower(r)
				default:
					return -1
				}
			}, string(n.Text(source)))
			id := slug
			for i := 1; headings[id]; i++ {
				id = fmt.Sprintf("%s-%d", slug, i)
			}
			headings[id], d.anchors[id] = true, true
		case *ast.RawHTML:
			for _, match := range htmlAnchor.FindAllSubmatch(n.Segments.Value(source), -1) {
				d.anchors[string(match[1])] = true
			}
		case *ast.HTMLBlock:
			for _, match := range htmlAnchor.FindAllSubmatch(n.Lines().Value(source), -1) {
				d.anchors[string(match[1])] = true
			}
		}
		return ast.WalkContinue, nil
	})
	return d
}

func localTarget(file string, u *url.URL) string {
	if u.Path == "" {
		return file
	}
	if strings.HasPrefix(u.Path, "/") {
		return filepath.Clean(strings.TrimLeft(u.Path, "/"))
	}
	return filepath.Join(filepath.Dir(file), filepath.FromSlash(u.Path))
}

func localError(file string, u *url.URL, docs map[string]document) error {
	target := localTarget(file, u)
	if target == ".." || strings.HasPrefix(target, ".."+string(filepath.Separator)) {
		return fmt.Errorf("target is outside the repository")
	}
	if _, err := os.Stat(target); err != nil {
		return err
	}
	if u.Fragment == "" || !strings.EqualFold(filepath.Ext(target), ".md") {
		return nil
	}
	d, ok := docs[target]
	if !ok {
		source, err := os.ReadFile(target)
		if err != nil {
			return err
		}
		d = parse(source)
		docs[target] = d
	}
	if !d.anchors[u.Fragment] {
		return fmt.Errorf("anchor %q not found in %s", u.Fragment, target)
	}
	return nil
}

var redirectLoop = fmt.Errorf("redirect loop or limit exceeded")

func checkHTTP(client *http.Client, address string, sleep func(time.Duration)) (string, string) {
	for attempt := 0; ; attempt++ {
		resp, err := client.Get(address)
		if err != nil {
			if e, ok := err.(*url.Error); ok && e.Err == redirectLoop {
				return "broken", redirectLoop.Error()
			}
			if attempt == 2 {
				return "unverified", err.Error()
			}
			sleep(time.Second << attempt)
			continue
		}
		_ = resp.Body.Close()
		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return "ok", ""
		case resp.StatusCode == 404 || resp.StatusCode == 410:
			return "broken", resp.Status
		case attempt < 2 && (resp.StatusCode == 429 || resp.StatusCode >= 500):
			delay := time.Second << attempt
			retry := resp.Header.Get("Retry-After")
			if seconds, err := strconv.Atoi(retry); err == nil && seconds >= 0 {
				if seconds > 5 {
					return "unverified", resp.Status + " (Retry-After exceeds retry budget)"
				}
				delay = max(delay, time.Duration(seconds)*time.Second)
			} else if until, err := http.ParseTime(retry); err == nil {
				if wait := time.Until(until); wait > 5*time.Second {
					return "unverified", resp.Status + " (Retry-After exceeds retry budget)"
				} else {
					delay = max(delay, wait)
				}
			}
			sleep(delay)
		default:
			return "unverified", resp.Status
		}
	}
}

func run(files []string, external bool, out io.Writer) int {
	if len(files) == 0 {
		// Imported references contain links relative to their upstream sites.
		result, err := exec.Command("git", "ls-files", "-z", "--", "*.md",
			":(exclude)agents/plugins/**",
			":(exclude)agents/skills/go-gherkin-testing/references/**").Output()
		if err != nil {
			fmt.Fprintln(out, err)
			return 2
		}
		for path := range bytes.SplitSeq(result, []byte{0}) {
			if len(path) != 0 {
				files = append(files, string(path))
			}
		}
	}
	docs := map[string]document{}
	urls := map[string]string{}
	counts := map[string]int{}
	status := 0
	for _, file := range files {
		file = filepath.Clean(file)
		source, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(out, err)
			status = 2
			continue
		}
		d := parse(source)
		docs[file] = d
		for _, link := range d.links {
			u, err := url.Parse(link)
			if err != nil {
				if !external {
					fmt.Fprintf(out, "%s: %q: %v\n", file, link, err)
					status = max(status, 1)
				}
				continue
			}
			if strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https") || (u.Scheme == "" && u.Host != "") {
				if external {
					if u.Scheme == "" {
						u.Scheme = "https"
					}
					switch host := strings.ToLower(u.Hostname()); {
					case host == "localhost", strings.HasSuffix(host, ".localhost"), strings.HasSuffix(host, ".local"), host == "127.0.0.1", host == "::1", host == "example.com", host == "example.org", host == "example.net":
						continue
					}
					if u.User != nil {
						counts["unverified"]++
						fmt.Fprintf(out, "unverified: %s: URL contains credentials\n", file)
						continue
					}
					u.Fragment, u.RawFragment = "", ""
					if _, exists := urls[u.String()]; !exists {
						urls[u.String()] = file
					}
				}
				continue
			}
			if !external && u.Scheme == "" {
				if err := localError(file, u, docs); err != nil {
					fmt.Fprintf(out, "%s: %q: %v\n", file, link, err)
					status = max(status, 1)
				}
			}
		}
	}
	if external {
		client := &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return redirectLoop
			}
			for _, prev := range via {
				if req.URL.String() == prev.URL.String() {
					return redirectLoop
				}
			}
			return nil
		}}
		addresses := make([]string, 0, len(urls))
		for address := range urls {
			addresses = append(addresses, address)
		}
		sort.Strings(addresses)
		for _, address := range addresses {
			kind, detail := checkHTTP(client, address, time.Sleep)
			counts[kind]++
			if kind != "ok" {
				fmt.Fprintf(out, "%s: %s: %s (%s)\n", kind, urls[address], address, detail)
			}
			if kind == "broken" {
				status = max(status, 1)
			}
		}
		fmt.Fprintf(out, "External links: %d OK, %d broken, %d unverified\n", counts["ok"], counts["broken"], counts["unverified"])
	}
	return status
}

func main() {
	external := flag.Bool("external", false, "check external HTTP(S) links instead")
	flag.Parse()
	os.Exit(run(flag.Args(), *external, os.Stdout))
}
