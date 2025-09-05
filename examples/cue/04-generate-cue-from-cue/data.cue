// Example values of different types for export.

package main

import (
	"list"
)

input: *10 | uint @tag(input, type=int)

output: {
	inputData:            input
	inputDataTransformed: input + 1
	sequence: list.Repeat([0, 1], input)

	struct: [name=string]: foo: "bar"
	struct: {for i, _ in list.Repeat([0], input) {"input_\(i)": _}}
}
