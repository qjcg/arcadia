package expand

// Word is a word with a per-byte quoted mask. Mask[i] is true if Value[i] was
// single-quoted, double-quoted, or escaped in the source. A nil mask means
// every byte is unquoted.
type Word struct {
	Value string
	Mask  []bool
}

// Pipeline runs the full expansion pipeline on a single word:
// brace expansion → variable expansion → glob expansion.
// Returns the expanded words. For words without braces, returns [word].
func Pipeline(word string, mask []bool, opts GlobOptions, expandVar func(string, []bool) (string, []bool)) ([]string, error) {
	// Step 1: brace expansion
	expanded := ExpandMasked(Word{Value: word, Mask: mask})

	// Step 2: variable expansion (for each expanded word)
	if expandVar != nil {
		var result []Word
		for _, w := range expanded {
			val, m := expandVar(w.Value, w.Mask)
			result = append(result, Word{Value: val, Mask: m})
		}
		expanded = result
	}

	// Step 3: glob expansion
	var globbed []string
	for _, w := range expanded {
		matches, err := GlobWithOptions(w.Value, w.Mask, opts)
		if err != nil {
			return nil, err
		}
		globbed = append(globbed, matches...)
	}

	return globbed, nil
}

// ExpandCommand expands a command name and args through the expansion pipeline.
// Returns the expanded name and args. The name is the first expanded word.
func ExpandCommand(name string, nameMask []bool, args []string, argsMask [][]bool, opts GlobOptions, expandVar func(string, []bool) (string, []bool)) (string, []string, error) {
	// Expand the command name (brace + var expansion, no globbing)
	expandedName := ExpandMasked(Word{Value: name, Mask: nameMask})
	cmdName := name
	if len(expandedName) > 0 {
		cmdName = expandedName[0].Value
	}
	if expandVar != nil {
		val, _ := expandVar(cmdName, nameMask)
		cmdName = val
	}

	// Expand each arg
	var expandedArgs []string
	for i, arg := range args {
		var m []bool
		if i < len(argsMask) {
			m = argsMask[i]
		}
		expanded, err := Pipeline(arg, m, opts, expandVar)
		if err != nil {
			return "", nil, err
		}
		expandedArgs = append(expandedArgs, expanded...)
	}

	return cmdName, expandedArgs, nil
}
