package main

import (
	"flag"
	"fmt"
)

func main() {
	flagYoloMode := flag.Bool("y", false, "YOLO mode")
	flag.Parse()

	if *flagYoloMode {
		fmt.Println("TODO: YOLO mode")
	}
}
