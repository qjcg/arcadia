package scaffold

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
)

func PromptForVariables(vars []Variable, quiet bool) map[string]string {
	values := make(map[string]string, len(vars))
	if quiet {
		for _, v := range vars {
			values[v.Name] = v.Default
		}
		return values
	}

	fields := make([]huh.Field, 0, len(vars))
	answers := make(map[string]*string, len(vars))
	for _, v := range vars {
		title := v.Prompt
		if v.Required {
			title += " (required)"
		}
		value := v.Default
		if v.Type == TypeChoice && len(v.Choices) > 0 {
			options := make([]huh.Option[string], len(v.Choices))
			for i, choice := range v.Choices {
				options[i] = huh.NewOption(choice, choice)
			}
			if !slices.Contains(v.Choices, value) {
				value = v.Choices[0]
			}
			fields = append(fields, huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&value))
		} else {
			input := huh.NewInput().Title(title).Value(&value)
			if v.Required {
				input.Validate(func(inputValue string) error {
					if strings.TrimSpace(inputValue) == "" {
						return errors.New("this value is required")
					}
					return nil
				})
			}
			fields = append(fields, input)
		}
		answers[v.Name] = &value
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, "Cancelled.")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
	for name, answer := range answers {
		values[name] = *answer
	}
	return values
}
