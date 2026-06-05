package main

import (
	"fmt"
	"math/rand"
)

func main() {
	pangrams := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Pack my box with five dozen liquor jugs.",
		"Sphinx of black quartz, judge my vow!",
		"How vexingly quick daft zebras jump!",
		"The five boxing wizards jump quickly.",
		"Jackdaws love my big sphinx of quartz.",
	}

	randomPangram := pangrams[rand.Intn(len(pangrams))]
	fmt.Println(randomPangram)
}
