package main

import (
	"os"

	"github.com/qjcg/arcadia/exp/tpi/cmd"
)

func main() {
	rootCmd := cmd.RootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}