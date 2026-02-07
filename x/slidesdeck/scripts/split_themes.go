package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	data, err := os.ReadFile("assets/src/css/vendor/themes.css")
	if err != nil {
		fmt.Printf("Error reading themes.css: %v\n", err)
		os.Exit(1)
	}

	content := string(data)

	// Pre-process: daisyUI v5 themes are concatenated.
	// We look for [data-theme=NAME]{...}
	// regexp to find theme name and content
	re := regexp.MustCompile(`\[data-theme=([a-z0-9-]+)\]\{(.*?)\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		fmt.Println("No themes found in themes.css")
		return
	}

	err = os.MkdirAll("assets/src/css/themes", 0o755)
	if err != nil {
		fmt.Printf("Error creating themes directory: %v\n", err)
		os.Exit(1)
	}

	for _, match := range matches {
		name := match[1]
		body := match[2]

		css := fmt.Sprintf("[data-theme=%s]{\n  %s\n}", name, strings.ReplaceAll(body, ";", ";\n  "))

		filename := fmt.Sprintf("assets/src/css/themes/%s.css", name)
		err = os.WriteFile(filename, []byte(css+"\n"), 0o644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("Extracted theme: %s\n", name)
	}
}
