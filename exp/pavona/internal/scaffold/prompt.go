package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PromptForVariables asks the user for each variable interactively.
// In quiet mode, returns defaults for everything.
func PromptForVariables(vars []Variable, quiet bool) map[string]string {
	values := make(map[string]string, len(vars))

	for _, v := range vars {
		if quiet {
			values[v.Name] = v.Default
			continue
		}

		prompt := v.Prompt
		if prompt == "" {
			prompt = v.Name
		}

		defaultStr := v.Default
		if defaultStr == "" {
			defaultStr = "(no default)"
		}

		reader := bufio.NewReader(os.Stdin)

		for {
			if len(v.Choices) > 0 {
				fmt.Printf("\n%s\n", prompt)
				for i, c := range v.Choices {
					mark := " "
					if c == v.Default {
						mark = "*"
					}
					fmt.Printf("  %s %d) %s\n", mark, i+1, c)
				}
				fmt.Printf("Enter number [1-%d] (default: %s): ", len(v.Choices), defaultStr)
			} else if v.Help != "" {
				fmt.Printf("%s [%s] (%s): ", prompt, defaultStr, v.Help)
			} else {
				fmt.Printf("%s [%s]: ", prompt, defaultStr)
			}

			raw, _ := reader.ReadString('\n')
			raw = strings.TrimSpace(raw)

			if raw == "" {
				values[v.Name] = v.Default
				break
			}

			if len(v.Choices) > 0 {
				if idx, err := strconv.Atoi(raw); err == nil && idx >= 1 && idx <= len(v.Choices) {
					values[v.Name] = v.Choices[idx-1]
					break
				}
				// Check if raw matches a choice name
				found := false
				for _, c := range v.Choices {
					if strings.EqualFold(raw, c) {
						values[v.Name] = c
						found = true
						break
					}
				}
				if found {
					break
				}
				fmt.Printf("Invalid choice. Please enter a number 1-%d.\n", len(v.Choices))
				continue
			}

			if v.Required && raw == "" {
				fmt.Println("This value is required.")
				continue
			}

			values[v.Name] = raw
			break
		}
	}

	return values
}
