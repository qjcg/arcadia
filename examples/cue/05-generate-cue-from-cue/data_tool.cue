package main

import (
	"encoding/json"
	"tool/file"
)

command: gen: file.Create & {
	$short:   "An alternative way to generate CUE from CUE."
	filename: "generated.cue"
	contents: """
	package main

	generated: \(json.Marshal(output))

	"""
}
