input := 20
out := generated.cue

.PHONY: gen
gen:
	cue eval -e output -s -t input=$(input) > $(out)

.PHONY: clean
clean:
	rm -rf $(out)
