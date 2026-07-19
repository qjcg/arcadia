package expand

// Pipeline runs the full expansion pipeline on a single word:
// brace expansion → variable expansion → glob expansion.
// Returns the expanded words. For words without braces, returns [word].
func Pipeline(word string, expandVar func(string) string) []string {
	// Step 1: brace expansion
	expanded := Expand(word)

	// Step 2: variable expansion (for each expanded word)
	if expandVar != nil {
		var result []string
		for _, w := range expanded {
			result = append(result, expandVar(w))
		}
		expanded = result
	}

	// Step 3: glob expansion
	var globbed []string
	for _, w := range expanded {
		globbed = append(globbed, GlobExpand(w)...)
	}

	return globbed
}

// ExpandCommand expands a command name and args through the expansion pipeline.
// Returns the expanded name and args. The name is the first expanded word.
func ExpandCommand(name string, args []string, expandVar func(string) string) (string, []string) {
	// Expand the command name
	expandedName := Expand(name)
	cmdName := name
	if len(expandedName) > 0 {
		cmdName = expandedName[0]
	}
	if expandVar != nil {
		cmdName = expandVar(cmdName)
	}

	// Expand each arg
	var expandedArgs []string
	for _, arg := range args {
		expanded := Pipeline(arg, expandVar)
		expandedArgs = append(expandedArgs, expanded...)
	}

	return cmdName, expandedArgs
}
