package main

import (
	"encoding/json"
	"encoding/toml"
	"encoding/yaml"
	"tool/exec"
	"tool/file"
)

command: gen: {
	_prefix: "generated"

	genCUE: exec.Run & {
		$short:   "Generate a CUE file from CUE."
		cmd: "cue export data.cue -p main -e output --outfile \(_prefix).cue --force"
	}

	genJSON: file.Create & {
		$short:   "Generate a JSON file from CUE."
		filename: "\(_prefix).json"
		contents: json.Marshal(output)
	}

	genYAML: file.Create & {
		$short:   "Generate a YAML file from CUE."
		filename: "\(_prefix).yaml"
		contents: yaml.Marshal(output)
	}

	genTOML: file.Create & {
		$short:   "Generate a TOML file from CUE."
		filename: "\(_prefix).toml"
		contents: toml.Marshal(output)
	}
}
