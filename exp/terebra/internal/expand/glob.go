package expand

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GlobOptions controls glob expansion behavior.
type GlobOptions struct {
	// NullGlob returns no words (an empty result) when a pattern matches
	// nothing, instead of returning the literal pattern.
	NullGlob bool
	// DotGlob includes filenames beginning with '.' in matches.
	DotGlob bool
}

// Glob expands glob patterns in a word using filepath.Match.
// Supports *, ?, [...], and ** (recursive).
// Returns the word unchanged if no match found or no glob chars present.
// mask marks quoted bytes; glob metacharacters inside quoted regions do not
// trigger expansion.
func Glob(word string, mask []bool) ([]string, error) {
	return GlobWithOptions(word, mask, GlobOptions{})
}

// GlobWithOptions expands a glob pattern with the given options.
func GlobWithOptions(word string, mask []bool, opts GlobOptions) ([]string, error) {
	if !hasUnquotedGlobChars(word, mask) {
		return []string{word}, nil
	}

	matches, err := globMatch(word, opts)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		if opts.NullGlob {
			return nil, nil
		}
		return []string{word}, nil
	}
	return matches, nil
}

// GlobExpand expands a word through the glob pipeline.
// If the word contains glob characters and matches files, returns matches.
// Otherwise returns the original word.
func GlobExpand(word string, mask []bool) ([]string, error) {
	return Glob(word, mask)
}

// hasUnquotedGlobChars reports whether s contains a glob metacharacter that is
// not inside a quoted region. A nil mask means every byte is unquoted.
func hasUnquotedGlobChars(s string, mask []bool) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*', '?', '[':
			if mask == nil || !mask[i] {
				return true
			}
		}
	}
	return false
}

// hasGlobChars reports whether s contains any glob metacharacter, quoted or not.
func hasGlobChars(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*', '?', '[':
			return true
		}
	}
	return false
}

// globMatch expands a pattern. It uses filepath.Glob for the common case and a
// custom walker when ** or dotglob semantics are required.
func globMatch(pattern string, opts GlobOptions) ([]string, error) {
	if !opts.DotGlob && !strings.Contains(pattern, "**") {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		// filepath.Glob matches dotfiles; exclude them unless the pattern
		// explicitly targets a leading dot.
		if !patternTargetsDot(pattern) {
			filtered := matches[:0]
			for _, m := range matches {
				if strings.HasPrefix(filepath.Base(m), ".") {
					continue
				}
				filtered = append(filtered, m)
			}
			matches = filtered
		}
		sort.Strings(matches)
		return matches, nil
	}
	return globWalk(pattern, opts)
}

// patternTargetsDot reports whether the final pattern segment begins with a
// literal dot, meaning dotfiles are explicitly targeted.
func patternTargetsDot(pattern string) bool {
	segs := strings.Split(pattern, "/")
	return strings.HasPrefix(segs[len(segs)-1], ".")
}

// globWalk walks the filesystem from the pattern's static root and matches each
// relative path against the pattern, supporting ** and dotglob.
func globWalk(pattern string, opts GlobOptions) ([]string, error) {
	if err := validatePattern(pattern); err != nil {
		return nil, err
	}
	root, rel := splitRoot(pattern)
	if _, err := os.Stat(root); err != nil {
		return nil, nil
	}

	var results []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath := path
		if root != "." {
			relPath = strings.TrimPrefix(path, root+"/")
		}
		if relPath == "." || relPath == "" {
			return nil
		}
		ok, err := matchGlob(rel, relPath, opts.DotGlob)
		if err != nil {
			return err
		}
		if ok {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(results)
	return results, nil
}

// splitRoot splits a pattern into a static directory root and the remaining
// pattern relative to that root. The root is the longest leading run of
// segments that contain no glob metacharacters.
func splitRoot(pattern string) (root, rel string) {
	segs := strings.Split(pattern, "/")
	firstGlob := -1
	for i, seg := range segs {
		if hasGlobChars(seg) {
			firstGlob = i
			break
		}
	}
	if firstGlob == -1 {
		return pattern, ""
	}
	root = strings.Join(segs[:firstGlob], "/")
	if root == "" {
		root = "."
	}
	rel = strings.Join(segs[firstGlob:], "/")
	return root, rel
}

// validatePattern returns an error if any non-** segment is a malformed glob.
func validatePattern(pattern string) error {
	for seg := range strings.SplitSeq(pattern, "/") {
		if seg == "**" {
			continue
		}
		if _, err := filepath.Match(seg, ""); err != nil {
			return err
		}
	}
	return nil
}

// matchGlob reports whether name matches pattern, where ** matches zero or more
// path segments and * ? [...] match within a single segment.
func matchGlob(pattern, name string, dotGlob bool) (bool, error) {
	return matchSegments(splitSegments(pattern), splitSegments(name), dotGlob)
}

func splitSegments(p string) []string {
	segs := strings.Split(p, "/")
	out := segs[:0]
	for _, s := range segs {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func matchSegments(pat, name []string, dotGlob bool) (bool, error) {
	if len(pat) == 0 {
		return len(name) == 0, nil
	}
	if pat[0] == "**" {
		for i := 0; i <= len(name); i++ {
			ok, err := matchSegments(pat[1:], name[i:], dotGlob)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	if len(name) == 0 {
		return false, nil
	}
	ok, err := matchSegment(pat[0], name[0], dotGlob)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return matchSegments(pat[1:], name[1:], dotGlob)
}

// matchSegment matches a single path segment against a single pattern segment.
func matchSegment(pattern, name string, dotGlob bool) (bool, error) {
	if !dotGlob && strings.HasPrefix(name, ".") && !strings.HasPrefix(pattern, ".") {
		return false, nil
	}
	if !dotGlob {
		return filepath.Match(pattern, name)
	}
	return matchSegmentDot(pattern, name)
}

// matchSegmentDot matches a single segment allowing * and ? to match a leading
// dot (dotglob semantics).
func matchSegmentDot(p, n string) (bool, error) {
	for len(p) > 0 {
		switch p[0] {
		case '*':
			for len(p) > 0 && p[0] == '*' {
				p = p[1:]
			}
			for i := 0; i <= len(n); i++ {
				ok, err := matchSegmentDot(p, n[i:])
				if err != nil {
					return false, err
				}
				if ok {
					return true, nil
				}
			}
			return false, nil
		case '?':
			if len(n) == 0 {
				return false, nil
			}
			p = p[1:]
			n = n[1:]
		case '[':
			matched, rest, err := matchClass(p, n)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
			p = rest
			n = n[1:]
		default:
			if len(n) == 0 || n[0] != p[0] {
				return false, nil
			}
			p = p[1:]
			n = n[1:]
		}
	}
	return len(n) == 0, nil
}

// matchClass matches the leading character class of p (which starts with '[')
// against the first byte of n. It returns whether it matched, the remaining
// pattern after the class, and any error.
func matchClass(p, n string) (matched bool, rest string, err error) {
	if len(n) == 0 {
		return false, "", nil
	}
	j := 1
	negate := false
	if j < len(p) && (p[j] == '^' || p[j] == '!') {
		negate = true
		j++
	}
	start := j
	for j < len(p) && p[j] != ']' {
		j++
	}
	if j >= len(p) {
		// Unclosed [ is treated as a literal.
		return false, "", nil
	}
	class := p[start:j]
	rest = p[j+1:]
	ch := n[0]
	inClass := false
	for k := 0; k < len(class); k++ {
		if k+2 < len(class) && class[k+1] == '-' {
			if ch >= class[k] && ch <= class[k+2] {
				inClass = true
			}
			k += 2
		} else if class[k] == ch {
			inClass = true
		}
	}
	if negate {
		inClass = !inClass
	}
	return inClass, rest, nil
}
