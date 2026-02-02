package main

import (
	"fmt"

	"github.com/qjcg/arcadia/x/elbereth/examples/packages/lib"
)

func main() {
	fmt.Println("Square of 5 is:", lib.Square(int64(5)))
	fmt.Println("Cube of 5 is:", lib.Cube(int64(5)))
}
