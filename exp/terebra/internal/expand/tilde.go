package expand

import (
	"os"
	"os/user"
	"strings"
)

// Tilde expands a leading unquoted ~ in w to the home directory.
// A bare ~ or ~/... expands to $HOME; ~user or ~user/... expands to that
// user's home directory. Returns w unchanged when no expansion applies
// (quoted ~, unset $HOME, or unknown user).
func Tilde(w Word) Word {
	if len(w.Value) == 0 || w.Value[0] != '~' {
		return w
	}
	// A quoted ~ is never expanded. A nil mask means every byte is unquoted.
	if w.Mask != nil && w.Mask[0] {
		return w
	}

	// Determine the home directory and how much of the prefix it consumes.
	home, consumed := tildeHome(w.Value)
	if home == "" {
		return w
	}

	rest := w.Value[consumed:]
	// Mark the inserted home bytes as quoted so glob metacharacters inside the
	// home path are not re-globbed (matches bash).
	homeMask := make([]bool, len(home))
	for i := range homeMask {
		homeMask[i] = true
	}
	var restMask []bool
	if w.Mask == nil {
		// Original word fully unquoted; the rest stays unquoted.
		restMask = make([]bool, len(rest))
	} else {
		restMask = w.Mask[consumed:]
	}
	return Word{
		Value: home + rest,
		Mask:  concatMasks(homeMask, restMask),
	}
}

// tildeHome returns the home directory for the tilde prefix in s and the
// number of leading bytes of s that the prefix consumes. s must start with
// '~'. Returns ("", 0) when no expansion applies.
func tildeHome(s string) (string, int) {
	rest := s[1:]
	switch {
	case rest == "" || rest[0] == '/':
		// ~ or ~/...
		return os.Getenv("HOME"), 1
	case rest[0] == '+' || rest[0] == '-':
		// ~+ / ~- (PWD / OLDPWD) are out of scope; leave literal.
		return "", 0
	}

	// ~user or ~user/... — consume the username up to the next '/' or end.
	end := strings.IndexByte(rest, '/')
	if end == -1 {
		end = len(rest)
	}
	name := rest[:end]
	u, err := user.Lookup(name)
	if err != nil {
		return "", 0
	}
	return u.HomeDir, 1 + end
}
