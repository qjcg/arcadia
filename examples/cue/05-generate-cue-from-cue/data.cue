// Example values of different types for export.

package main

input: *10 | uint @tag(input, type=int)

output: {
	inputData:            input
	inputDataTransformed: input + 1
	sequence:             input * [0, 1]

	struct: [name=string]: foo: "bar"
	struct: {for i, _ in input * [0] {"input_\(i)": _}}
}
