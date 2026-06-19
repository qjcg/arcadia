package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type validateCCParams struct {
	File      []string `descr:"File(s) containing commit messages to validate" short:"f" optional:"true"`
	MaxLength int      `descr:"Max length of the subject line" short:"m" optional:"true"`
}

func createValidateCCCmd() *cobra.Command {
	return boa.CmdT[validateCCParams]{
		Use:   "validate-cc",
		Short: "Validate conventional commit messages",
		Long: "Validate commit messages against the Conventional Commits specification.\n\n" +
			"Reads commit messages from stdin (one per invocation) or from one or more files via --file.\n" +
			"Use --max-length (-m) to enforce a maximum subject line length.",
		RunFuncE: func(p *validateCCParams, cmd *cobra.Command, _ []string) error {
			return runValidateCCCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runValidateCCCmd(p *validateCCParams, cmd *cobra.Command) error {
	var messages []string

	if len(p.File) > 0 {
		for _, f := range p.File {
			data, err := os.ReadFile(f)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", f, err)
			}
			messages = append(messages, string(data))
		}
	} else {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("error reading stdin: %w", err)
		}
		messages = append(messages, string(data))
	}

	hasErrors := false
	for _, msg := range messages {
		errs := validateConventionalCommit(msg, p.MaxLength)
		if len(errs) > 0 {
			hasErrors = true
			fmt.Fprintln(cmd.OutOrStdout(), "INVALID")
			for _, e := range errs {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "VALID")
		}
	}

	if hasErrors {
		return fmt.Errorf("one or more commit messages are invalid")
	}
	return nil
}

// validateConventionalCommit checks a commit message against the Conventional Commits spec.
// Returns a list of validation errors, empty if the message is valid.
func validateConventionalCommit(msg string, maxLength int) []string {
	var errs []string

	// Normalize line endings
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.TrimSpace(msg)

	if msg == "" {
		return []string{"commit message is empty"}
	}

	// Subject is the first line
	lines := strings.SplitN(msg, "\n", 2)
	subject := strings.TrimSpace(lines[0])

	// Check max length
	if maxLength > 0 && len(subject) > maxLength {
		errs = append(errs, fmt.Sprintf(
			"subject line is %d characters, exceeds max length of %d",
			len(subject), maxLength,
		))
	}

	// Check conventional commit format: type[(scope)][!]: description
	// Description must be non-empty (at least one non-whitespace character after ": ")
	re := regexp.MustCompile(`^[a-zA-Z]+(\([^)]*\))?!?:\s+\S`)
	if !re.MatchString(subject) {
		errs = append(errs, "does not follow conventional commit format (type[(scope)][!]: description)")
	}

	return errs
}
